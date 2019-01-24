package broker

import (
	//"errors"
	//"io"
	. "github.com/pobearm/MsgBroker/packets"
	"net"
	"sync"
	"time"
)

type Client struct {
	waitGroup        sync.WaitGroup     //关闭时, 等待多任务完成
	clientID         string             //clientid的
	conn             net.Conn           //连接
	keepAlive        uint16             //心跳间隔时间(秒)
	TimeOutTime      int64              //客户端超时的时间(utc秒)
	ConnectTime      time.Time          //连入的时间
	state            *State             //连接状态
	outboundMessages chan ControlPacket //要发出去的消息chan
	stopChan         chan struct{}      //内部的client关闭通知channel
	// handleChan       chan *HandleMsg    //client对外通知channnel
	tempStopOnce  *sync.Once //暂时关闭操作的时候, 锁住保证只有一次调用.
	totalStopOnce *sync.Once //完全关闭操作的时候, 锁住保证只有一次调用.
	takeOverLock  sync.Mutex //
	idGenerator   *IdGen     //内部消息id生成器
}

// func newClient(clientID string, maxQDepth int, handleChan chan *HandleMsg) *Client {
func newClient(clientID string, maxQDepth int) *Client {
	return &Client{
		waitGroup:        sync.WaitGroup{},
		clientID:         clientID,
		outboundMessages: make(chan ControlPacket, maxQDepth),
		totalStopOnce:    new(sync.Once),
		idGenerator:      NewIdGen(),
		// handleChan:       handleChan,
		state: new(State),
	}
}

// func (c *Client) Start(alive uint16, persi Persistence, conn net.Conn) {
func (c *Client) Start(alive uint16, conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			// log.Errorf("client.Start", "client painc: %#v", err)
		}
	}()
	defer c.takeOverLock.Unlock()
	c.takeOverLock.Lock()
	c.ConnectTime = time.Now()
	//如果这个clientid已经处于连接状态,就先关闭他
	if c.Connected() {
		c.TempStop()
	}
	//重用client对象
	c.state.SetValue(CONNECTING)
	c.conn = conn
	//给一个一次锁
	c.tempStopOnce = new(sync.Once)
	c.stopChan = make(chan struct{})
	//客户端的心跳间隔
	c.keepAlive = alive
	//计算超时时间
	c.resetTimeOut()
	//发回执
	ca := NewControlPacket(CONNACK).(*ConnackPacket)
	ca.ReturnCode = CONN_ACCEPTED
	ca.Write(c.conn)
	//开发接收和发送
	c.waitGroup.Add(2)
	go c.Receive()
	go c.Send()
	//标记状态
	c.state.SetValue(CONNECTED)
	//装载离线消息
	// c.LoadLeaveMsg(persi)
	// log.Debug("client.Start", "a client starting ok:", c.clientID)
}

//加载离线消息
// func (c *Client) LoadLeaveMsg(persi Persistence) {
// 	//log.Debug("load leave message")
// 	//如果不存在消息持久对象, 新开一个
// 	if !persi.Exists(c.clientID) {
// 		persi.Open(c.clientID)
// 	}

// 	msgs := persi.GetAll(c.clientID)
// 	for _, msg := range msgs {
// 		switch msg.(type) {
// 		case *PublishPacket:
// 			//有可能消息已经推送过了, 但是没收到回执, 所以没有删除
// 			pp := msg.(*PublishPacket)
// 			if pp.Qos > 0 {
// 				//表示可能是重复的消息
// 				pp.Dup = true
// 			}
// 			c.outboundMessages <- pp
// 		default:
// 			c.outboundMessages <- msg
// 		}
// 	}
// 	//log.Debug("load leave message finished")
// }

func (c *Client) Receive() {
	for {
		select {
		case <-c.stopChan:
			c.waitGroup.Done()
			return
		default:
			cp, err := ReadPacket(c.conn)
			if err != nil {
				// log.Error("client.Receive", c.clientID, ": in receive: ", err.Error())
				c.waitGroup.Done()
				if c.Connected() {
					go c.TempStop()
				}
				return
			} else {
				//收到任何消息, 都会重置超时时间
				c.resetTimeOut()
				//处理各种消息
				c.handleRecieve(cp)
			}
		}
	}
}

func (c *Client) handleRecieve(cp ControlPacket) {
	switch cp.(type) {
	case *ConnectPacket:
		//多次收到连接请求是协议冲突的, 忽略, 什么也不做
		// log.Debug("client.handleRecieve", "get another connect pack client: ", c.clientID)
	case *DisconnectPacket:
		// log.Debug("client.handleRecieve", "get disconnect pack: ", c.clientID)
		//客户端主动断开, 也可以暂时不回收资源.
		go c.TempStop()
	case *PubackPacket:
		//收到回执
		pa := cp.(*PubackPacket)
		//释放消息id
		c.idGenerator.FreeId(pa.MessageID)
		//把消息放入到chan, 由broker处理
		// msg := &HandleMsg{
		// 	MsgType:  "pubback",
		// 	ClientId: c.clientID,
		// 	Cp:       cp,
		// }
		// c.handleChan <- msg
	case *PubrecPacket:
		//qos=2 会收到这个, 服务端要发送回执
		pr := cp.(*PubrecPacket)
		prel := NewControlPacket(PUBREL).(*PubrelPacket)
		prel.MessageID = pr.MessageID
		c.handleFlow(prel)
	case *PubrelPacket:
		//We received a PUBREL for a QoS2 PUBLISH from the client, hrotti delivers on PUBLISH though
		//so we've already sent the original message to any subscribers, so just create a new
		//PUBCOMP message with the correct message id and pass it to the handleFlow function.
		pr := cp.(*PubrelPacket)
		pc := NewControlPacket(PUBCOMP).(*PubcompPacket)
		pc.MessageID = pr.MessageID
		c.handleFlow(pc)
	case *PubcompPacket:
		//Received a PUBCOMP for a QoS2 PUBLISH we originally sent the client. Check the messageid is
		//one we think is in use, if so delete the original PUBLISH from the outbound persistence store
		//and free the message id for reuse
		pc := cp.(*PubcompPacket)
		//释放消息id
		c.idGenerator.FreeId(pc.MessageID)
		//把消息放入到chan, 由broker处理
		// msg := &HandleMsg{
		// 	MsgType:  "pubcomp",
		// 	ClientId: c.clientID,
		// 	Cp:       cp,
		// }
		// c.handleChan <- msg
	case *PingreqPacket:
		//ping
		presp := NewControlPacket(PINGRESP).(*PingrespPacket)
		//log.Debug("get a ping from: ", c.clientID)
		c.handleFlow(presp)
	}
}

func (c *Client) handleFlow(msg ControlPacket) {
	select {
	case c.outboundMessages <- msg:
	default:
		// log.Errorf("client.handleFlow", "client %s out chan full, message droped", c.clientID)
		//消息队列满了,通常不应该出现
	}
}

func (c *Client) Send() {
	for {
		select {
		case <-c.stopChan:
			//log.Debug("send out: ", c.clientID)
			c.waitGroup.Done()
			return
		case msg := <-c.outboundMessages:
			if msg == nil {
				// log.Error("client.Send", "msg is nil")
				continue
			}
			if c.conn == nil {
				// todo: 目前丢弃了消息, 应该重发?
				// log.Error("client.Send", "conn is nil")
				continue
			}
			msg.Write(c.conn)
		}
	}
}

//客户端完全停止,从clients中调用: keepalive超时.
//client对象会从内存中删除.
func (c *Client) TotalStop() {
	// log.Debug("client.TotalStop", "get in TotalStop: ", c.clientID)
	c.totalStopOnce.Do(func() {
		//先做个内部停止
		c.TempStop()
		//关闭发送消息的channnel
		close(c.outboundMessages)
		//告知broker, client已经从外部停止了
		// msg := &HandleMsg{
		// 	MsgType:  "stop",
		// 	ClientId: c.clientID,
		// }
		// c.handleChan <- msg
		// log.Debug("client.TotalStop", "TotalStoped client: ", c.clientID)
	})
}

//临时关闭, client仍然在内存中,可以被重用. 触发的情况是:
//1.clinets检查到TimeOutTime过期了
//2. 客户端通过mqtt协议主动断开
//3. 网络异常断开
func (c *Client) TempStop() {
	c.tempStopOnce.Do(func() {
		// log.Debug("client.TempStop", "get in tempStopOnce: ", c.clientID)
		//if c.Connected() {
		c.state.SetValue(DISCONNECTING)
		//发通知给receive 和send, 退出循环
		close(c.stopChan)
		//关闭连接
		c.conn.Close()
		c.conn = nil
		//等待recieve和send两个循环都结束
		c.waitGroup.Wait()
		//标记为已经断开连接
		c.state.SetValue(DISCONNECTED)
		// log.Debug("client.TempStop", "tempStoped client: ", c.clientID)
		//}
	})
}

//设置超时时间
func (c *Client) resetTimeOut() {
	c.TimeOutTime = time.Now().Unix() + int64(int32(c.keepAlive)+30)
	//log.Debug("reset timeout: ", c.TimeOutTime)
}

//检查是否允许客户端重连
func (c *Client) OutTime() bool {
	ts := c.ConnectTime.Add(10*time.Millisecond).Unix() - time.Now().Unix()
	// ts := c.ConnectTime.Add(time.Duration(c.keepAlive)*time.Second).Unix() - time.Now().Unix()
	if ts > 0 {
		return false
	} else {
		return true
	}
}

//是否是连接状态
func (c *Client) Connected() bool {
	return c.state.Value() == CONNECTED
}

//是否是断开连接状态
func (c *Client) DisConnected() bool {
	return c.state.Value() == DISCONNECTED
}
