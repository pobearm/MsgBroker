package broker

import (
	//"code.google.com/p/go.net/websocket"
	//"net/http"
	//"net/url"
	//"errors"
	"fmt"
	. "github.com/pobearm/MsgBroker/packets"
	"net"
	"time"
)

type HandleMsg struct {
	MsgType  string // publish, sub, unsub, stop
	ClientId string
	Cp       ControlPacket
}

//服务对象
type Broker struct {
	tcpListener   net.Listener //tcp
	maxQueueDepth int          //最大消息数
	// PersistStore  Persistence  //持久化接口
	// HandleChan   chan *HandleMsg //接收来自client的通知,并进行处理
	clientsWheel *TimeWheel //连接的客户端
	Username     string
	Password     string
	BrokerStop   chan struct{} //broker stop
	//wsListener    net.Listener    //websocket
}

func NewBroker(maxDepth int, username string, password string) *Broker {
	//消息持久化模块
	// persist := &MemoryPersistence{}
	// persist.Init()

	b := &Broker{
		// PersistStore:  persist,
		// HandleChan:    make(chan *HandleMsg, 1000),
		maxQueueDepth: maxDepth,
		clientsWheel:  NewTimeWheel(1*time.Second, 100),
		Username:      username,
		Password:      password,
		BrokerStop:    make(chan struct{}),
	}
	return b
}

func (b *Broker) beginTcp(address string) error {
	//监听tcp端口  todo: ssl
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("open tcp %s failed: %s", address, err.Error())
	}
	b.tcpListener = ln
	// log.Info("broker.beginTcp", "start listen on ", address)
	go func() {
		for {
			select {
			case <-b.BrokerStop:
				break
			default:
				conn, err := ln.Accept() //接收
				if err != nil {
					//记录日志
					// log.Error("broker.beginTcp", "tcp accept error: ", err.Error())
					continue
				}
				//log.Debug("new incoming connect: ", conn.RemoteAddr())
				//初始化连接
				go b.initClient(conn)
			}
		}
	}()
	return nil
}

func (b *Broker) Start(tcp, wbsocket string) error {
	// log.Debug("broker.Start", "broker starting ...")
	//开启tcp
	if err := b.beginTcp(tcp); err != nil {
		return err
	}
	//开启websocket
	//b.beginWebsocket(wbsocket)
	//开始处理client的消息
	// go b.HandleClientMsg()
	// log.Debug("broker.Start", "starting finished.....")
	return nil
}

//初始化并启动监听
func (b *Broker) initClient(conn net.Conn) {
	rp, err := ReadPacket(conn)
	if err != nil {
		// log.Error("broker.initClient", "init client: ", err.Error())
		conn.Close()
		return
	}
	cp, ok := rp.(*ConnectPacket)
	if !ok {
		ca := NewControlPacket(CONNACK).(*ConnackPacket)
		ca.ReturnCode = CONN_NETWORK_ERROR
		ca.Write(conn)
		conn.Close()
		return
	}
	rc := cp.Validate(b.Username, b.Password)

	//如果没有验证通过
	if rc != CONN_ACCEPTED {
		if rc != CONN_PROTOCOL_VIOLATION {
			ca := NewControlPacket(CONNACK).(*ConnackPacket)
			ca.ReturnCode = rc
			ca.Write(conn)
		}
		// log.Errorf("broker.initClient", "connect err: %s, %v",
		// 	ConnackReturnCodes[rc], conn.RemoteAddr())
		conn.Close()
		return
	} else {
		//连接成功
		// log.Infof("broker.initClient", "connect success: %s, %v, %v",
		// 	ConnackReturnCodes[rc], cp.ClientIdentifier, conn.RemoteAddr())
	}
	//取客户端
	c, ok := b.clientsWheel.Find(cp.ClientIdentifier)
	if ok {
		//已经断开了, 重用已存在才client对象
		//而且上一次client的连接时间, 过了keeapalive值.
		if c.DisConnected() && c.OutTime() {
			//开始
			// c.Start(cp.KeepaliveTimer, b.PersistStore, conn)
			c.Start(cp.KeepaliveTimer, conn)
		} else { //连接中, 和断开连接中, 都拒绝
			// log.Errorf("broker.initClient", "client is handling now : %v", cp.ClientIdentifier)
			if conn != nil {
				ca := NewControlPacket(CONNACK).(*ConnackPacket)
				ca.ReturnCode = CONN_NETWORK_ERROR
				ca.Write(conn)
				conn.Close()
			}
			return
		}
	} else { //没有找到clinet对象, 新创建一个.
		// c = newClient(cp.ClientIdentifier, b.maxQueueDepth, b.HandleChan)
		c = newClient(cp.ClientIdentifier, b.maxQueueDepth)
		// log.Debug("broker.initClient", "create a new client: ", c.clientID)
		b.clientsWheel.AddClient(c.clientID, c)
		//开始
		// c.Start(cp.KeepaliveTimer, b.PersistStore, conn)
		c.Start(cp.KeepaliveTimer, conn)
	}
}

//停止listener
func (b *Broker) Stop() {
	// log.Debug("broker.Stop", "stoping server...")
	//关闭所有客户端
	b.clientsWheel.Stop()
	//关闭listener
	//b.wsListener.Close()
	//发送broker关闭信号
	close(b.BrokerStop)
	b.tcpListener.Close()
	// log.Debug("broker.Stop", "stoping listener end ....")
}

//处理来自client的通知
// func (b *Broker) HandleClientMsg() {
// 	for {
// 		select {
// 		case <-b.BrokerStop:
// 			return
// 		case msg := <-b.HandleChan:
// 			switch msg.MsgType {
// 			case "stop":
// 				//客户端关闭了, 只有从timewheel扫描时,从外部关闭才进入这里
// 				//此时, client已经从内存中移除了
// 				log.Infof("broker.HandleClientMsg", "client %s closed from out", msg.ClientId)
// 				//清理离线消息
// 				b.PersistStore.Close(msg.ClientId)
// 			case "pubback":
// 				//qos=1收到回执
// 				pa := msg.Cp.(*PubackPacket)
// 				b.PersistStore.Delete(msg.ClientId, pa.MessageID)
// 			case "pubcomp":
// 				//qos=2收到回执, 从持久化中删除这个消息
// 				pc := msg.Cp.(*PubcompPacket)
// 				b.PersistStore.Delete(msg.ClientId, pc.MessageID)
// 			default:
// 				log.Error("broker.HandleClientMsg", "unkown msg type: ", msg.MsgType)
// 			}
// 		}
// 	}
// }

/*
	注意: 现在直接把topic当作clientid来使用.
	不需要订阅或者取消订阅
*/
func (b *Broker) pushToOneClient(clientId string, pp *PublishPacket) {
	tc, ok := b.clientsWheel.Find(clientId)
	if !ok {
		// log.Error("broker.pushToOneClient", "not find client: ", clientId)
		//没有这个client, 消息就丢弃了.
		return
	}
	//先要给pack设置messageid, 一个客户端最多65535个id
	pp.MessageID = tc.idGenerator.GetMsgId()
	//如果qos>0, 就先写入持久化
	// 不支持qos==1，todo：
	// if pp.Qos > 0 {
	// 	b.PersistStore.Add(clientId, pp.MessageID, pp)
	// }
	//判断是否处于连接状态
	if tc.Connected() {
		tc.outboundMessages <- pp
		// log.Info("broker.pushToOneClient", "sended msg to client: ", tc.clientID)
	} else {
		// log.Info("broker.pushToOneClient", "client not connect: ", tc.clientID)
	}
}

func (b *Broker) Push(token string, qos byte, payload []byte) {
	if token == "" {
		// log.Error("broker.Push", "token is null or empty")
		return
	}
	pp := &PublishPacket{
		FixedHeader: FixedHeader{
			MessageType: PUBLISH,
			Qos:         qos,
		},
		TopicName: "",
		Payload:   payload,
	}
	//pp.SetRemainLength()
	go b.pushToOneClient(token, pp)
}

//func (b *Broker) beginWebsocket(address string) error {
//fmt.Println("websocket: ", wbsocket)
//监听websocet端口
//todo ssl
// lnwb, err := net.Listen("tcp", wbsocket)
// if err != nil {
// 	return errors.New(fmt.Sprintf("open tcp %s failed: %s", wbsocket, err.Error()))
// }
// var server websocket.Server
// server.Handshake = func(c *websocket.Config, req *http.Request) error {
// 	c.Origin, _ = url.Parse(req.RemoteAddr)
// 	c.Protocol = []string{"mqtt"}
// 	return nil
// }
// server.Handler = func(ws *websocket.Conn) {
// 	ws.PayloadType = websocket.BinaryFrame
// 	log.Info("new incoming websocket connect: %v", ws.RemoteAddr())
// 	b.InitClient(ws)
// }
// http.Handle("/", server)
// fmt.Println("start linsten wbsocket")
// b.wsListener = lnwb
// go func(ln net.Listener) {
// 	err := http.Serve(ln, nil)
// 	if err != nil {
// 		log.Error("%s")
// 	}
// }(lnwb)
// log.Info("start listen on %s", wbsocket)
//return nil
//}
