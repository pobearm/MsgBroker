package broker

import (
	. "github.com/pobearm/MsgBroker/packets"
	"sync"
)

// todo: 这个类的写法有问题， 字典取值有可能不存在。

//一个客户端的消息集合
//由于id最大65535个, 所以一个客户端能缓存的消息也只有这么多.
type MessageEntry struct {
	sync.Mutex
	//key,msgid, value: pack
	messages map[uint16]ControlPacket
}

//消息存储器
type MemoryPersistence struct {
	sync.RWMutex
	//输出的消息,key:clientid, value:message集合
	outbound map[string]*MessageEntry
}

func (p *MemoryPersistence) Init() error {
	//初始化
	p.outbound = make(map[string]*MessageEntry)
	return nil
}

func (p *MemoryPersistence) Open(client string) {
	p.Lock()
	defer p.Unlock()
	//增加一个客户端
	p.outbound[client] = &MessageEntry{messages: make(map[uint16]ControlPacket)}
}

func (p *MemoryPersistence) Close(client string) {
	p.Lock()
	defer p.Unlock()
	//移除一个客户端
	delete(p.outbound, client)
}

func (p *MemoryPersistence) Add(client string, msgid uint16, message ControlPacket) bool {
	//消息方向
	p.outbound[client].Lock()
	defer p.outbound[client].Unlock()
	//如果消息已经存在
	if _, ok := p.outbound[client].messages[msgid]; ok {
		return false
	}
	//加入消息
	p.outbound[client].messages[msgid] = message
	return true
}

func (p *MemoryPersistence) Delete(client string, msgid uint16) bool {
	//判断消息方向
	p.outbound[client].Lock()
	defer p.outbound[client].Unlock()
	if _, ok := p.outbound[client].messages[msgid]; !ok {
		return false
	}
	delete(p.outbound[client].messages, msgid)
	return true
}

func (p *MemoryPersistence) GetAll(client string) (messages []ControlPacket) {
	p.outbound[client].Lock()
	defer p.outbound[client].Unlock()
	//拿到所有的出的消息
	for _, message := range p.outbound[client].messages {
		messages = append(messages, message)
	}
	return messages
}

func (p *MemoryPersistence) Exists(client string) bool {
	p.RLock()
	defer p.RUnlock()
	_, okOutbound := p.outbound[client]
	return okOutbound
}
