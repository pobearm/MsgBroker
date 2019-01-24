package broker

import (
	. "github.com/pobearm/MsgBroker/packets"
)

type dirFlag byte

// const (
// 	INBOUND  = 1
// 	OUTBOUND = 2
// )

type Persistence interface {
	Init() error
	//clientId
	Open(string)
	//clientId
	Close(string)
	//clientId,  msgid,  pack
	Add(string, uint16, ControlPacket) bool
	//clientId,  msgid
	Delete(string, uint16) bool
	//clientId, out or in, pack
	GetAll(string) []ControlPacket
	//clientId
	Exists(string) bool
}
