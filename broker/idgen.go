package broker

import (
	"sync"
)

const (
	msgIDMax uint16 = 65535
	msgIDMin uint16 = 1
)

type IdGen struct {
	lock      sync.Mutex
	index     map[uint16]byte //已使用的索引
	now_index uint16          //当前索引
}

func NewIdGen() *IdGen {
	return &IdGen{
		lock:      sync.Mutex{},
		index:     make(map[uint16]byte),
		now_index: uint16(1),
	}
}

//生成一个id
func (i *IdGen) GetMsgId() uint16 {
	i.lock.Lock()
	defer i.lock.Unlock()
	c := uint16(0)
	for {
		if i.now_index == msgIDMax {
			i.now_index = 0
		}
		i.now_index++
		//如果没有被使用
		if _, ok := i.index[i.now_index]; !ok {
			i.index[i.now_index] = 1
			return i.now_index
		}
		//如果整个id库遍历完, 都没有可用的, 就使用当前index
		c++
		if c == msgIDMax {
			return i.now_index
		}
	}
}

//id是否被使用了
func (i *IdGen) InUse(id uint16) bool {
	i.lock.Lock()
	defer i.lock.Unlock()
	_, ok := i.index[id]
	return ok
}

//释放这个id
func (i *IdGen) FreeId(id uint16) {
	i.lock.Lock()
	defer i.lock.Unlock()
	_, ok := i.index[id]
	if ok {
		delete(i.index, id)
	}
}
