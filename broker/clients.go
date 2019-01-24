package broker

import (
	"sync"
	"time"
)

const (
	//默认的客户端在内存中保留的时间(秒), 24小时内有效
	CLIENT_ALIVE = 3600 * 12
)

//这里使用时间轮,
type clients struct {
	lock sync.Mutex
	list map[string]*Client //key: clientid
}

func NewClients() *clients {
	c := &clients{
		lock: sync.Mutex{},
		list: make(map[string]*Client),
	}
	return c
}

func (c *clients) Exist(id string) bool {
	_, ok := c.list[id]
	return ok
}

func (c *clients) Remove(id string) {
	c.lock.Lock()
	delete(c.list, id)
	c.lock.Unlock()
}

func (c *clients) Add(id string, cl *Client) {
	c.lock.Lock()
	c.list[id] = cl
	c.lock.Unlock()
}

func (c *clients) StopAll() {
	//要停止全部客户端, 不使用锁了.
	// log.Debug("clients.StopAll", "stop all client")
	for _, cl := range c.list {
		cl.TotalStop()
	}
	// log.Debug("clients.StopAll", "stop all finished")
}

//将超时的client断开
func (c *clients) ClearTimeOut() {
	// log.Debug("in clients ClearTimeOut")
	//c.lock.Lock()
	if len(c.list) == 0 {
		return
	}
	for id, cl := range c.list {
		nowt := time.Now().Unix()
		t := nowt - cl.TimeOutTime
		if cl.Connected() { //如果是连接状态
			if t >= 0 { //如果超时时间小于当前时间, 就认为超时了
				// log.Debug("clients.ClearTimeOut", "client timeout: ", cl.clientID, " key: ", id)
				//调用客户端的停用方法
				go cl.TempStop()
			}
		} else { //已经断开的客户端
			if t >= CLIENT_ALIVE { //断开时间已经超过最大内存保留时间
				//log.Debug("client removed: ", cl.clientID)
				go cl.TotalStop()
				//把客户端从内存中删除
				c.Remove(id)
			}
		}
	}
	//c.lock.Unlock()
}
