package broker

import (
	"sync"
)

type State struct {
	sync.RWMutex
	value StateVal
}

type StateVal uint8

const (
	DISCONNECTED  StateVal = 0x00 //断开连接了
	CONNECTING    StateVal = 0x01 //正在连入
	CONNECTED     StateVal = 0x02 //连接状态
	DISCONNECTING StateVal = 0x03 //正在断开中
	//TAKING        StateVal = 0x04 //正在重用操作中
)

func (s *State) SetValue(value StateVal) {
	s.Lock()
	s.value = value
	s.Unlock()
}

func (s *State) Value() StateVal {
	defer s.RUnlock()
	s.RLock()
	return s.value
}
