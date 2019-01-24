package broker

import (
	. "github.com/pobearm/MsgBroker/packets"
	"sync"
)

//保留消息字典
type RetainedMap struct {
	lock sync.RWMutex
	//key:topic, value: PublishPacket
	RetainedMessages map[string]*PublishPacket
}

func NewRetaindMap() *RetainedMap {
	return &RetainedMap{
		lock:             sync.RWMutex{},
		RetainedMessages: make(map[string]*PublishPacket),
	}
}

//设置保留消息, 一个topic只有一个
func (r *RetainedMap) SetRetained(topic string, message *PublishPacket) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if len(message.Payload) == 0 {
		delete(r.RetainedMessages, topic)
	} else {
		r.RetainedMessages[topic] = message
	}
}

//---------------------------------------------------

//topic与client的订阅关系
type TopicManager struct {
	lock sync.RWMutex
	//key: topic, value: clientids []string
	Topics map[string]map[string]byte
}

func NewTopicManager() *TopicManager {
	return &TopicManager{
		lock:   sync.RWMutex{},
		Topics: make(map[string]map[string]byte),
	}
}

//增加订阅
func (t *TopicManager) AddSub(clientid, topic string) {
	if clientid == "" || topic == "" {
		return
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	//topic exist
	cs, ok := t.Topics[topic]
	if !ok {
		cs = make(map[string]byte)
		t.Topics[topic] = cs
	}
	cs[clientid] = 1
}

//减少订阅
func (t *TopicManager) RemoveSub(clientid, topic string) {
	if clientid == "" || topic == "" {
		return
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	cs, ok := t.Topics[topic]
	//如果有记录才处理
	if !ok {
		return
	}
	delete(cs, clientid)
	//如果没有人订阅了, 就删除这个建
	if len(cs) == 0 {
		delete(t.Topics, topic)
	}
}

//移除一个客户端
func (t *TopicManager) RemoveClient(clientid string) {
	if clientid == "" {
		return
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	for topic, cs := range t.Topics {
		if _, ok := cs[clientid]; ok {
			delete(cs, clientid)
		}
		//如果没有人订阅了, 就删除这个建
		if len(cs) == 0 {
			delete(t.Topics, topic)
		}
	}
}
