package packets

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type ControlPacket interface {
	Write(io.Writer) error
	Unpack(io.Reader) error
	String() string
	Details() Details
	//Validate() bool
}

var PacketNames = map[uint8]string{
	1:  "CONNECT",
	2:  "CONNACK",
	3:  "PUBLISH",
	4:  "PUBACK",
	5:  "PUBREC",
	6:  "PUBREL",
	7:  "PUBCOMP",
	8:  "SUBSCRIBE",
	9:  "SUBACK",
	10: "UNSUBSCRIBE",
	11: "UNSUBACK",
	12: "PINGREQ",
	13: "PINGRESP",
	14: "DISCONNECT",
}

const (
	CONNECT     = 1
	CONNACK     = 2
	PUBLISH     = 3
	PUBACK      = 4
	PUBREC      = 5
	PUBREL      = 6
	PUBCOMP     = 7
	SUBSCRIBE   = 8
	SUBACK      = 9
	UNSUBSCRIBE = 10
	UNSUBACK    = 11
	PINGREQ     = 12
	PINGRESP    = 13
	DISCONNECT  = 14
)

const (
	CONN_ACCEPTED           = 0x00
	CONN_REF_BAD_PROTO_VER  = 0x01
	CONN_REF_ID_REJ         = 0x02
	CONN_REF_SERV_UNAVAIL   = 0x03
	CONN_REF_BAD_USER_PASS  = 0x04
	CONN_REF_NOT_AUTH       = 0x05
	CONN_NETWORK_ERROR      = 0xFE
	CONN_PROTOCOL_VIOLATION = 0xFF
)

var ConnackReturnCodes = map[uint8]string{
	0:   "Connection Accepted",
	1:   "Connection Refused: Bad Protocol Version",
	2:   "Connection Refused: Client Identifier Rejected",
	3:   "Connection Refused: Server Unavailable",
	4:   "Connection Refused: Username or Password in unknown format",
	5:   "Connection Refused: Not Authorised",
	254: "Connection Error",
	255: "Connection Refused: Protocol Violation",
}

func ReadPacket(r io.Reader) (cp ControlPacket, err error) {
	var fh FixedHeader
	b := make([]byte, 1)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	fh.unpack(b[0], r)
	// 解析包的长度字段时, 如果出错, 要处理, 不然会崩溃
	if fh.RemainingLength == -1 {
		return nil, errors.New("get fixheader.remainingLength error")
	}
	cp = NewControlPacketWithHeader(fh)
	if cp == nil {
		return nil, errors.New("Bad data from client")
	}
	//todo: 这里应该使用buf来重用内存
	packetBytes := make([]byte, fh.RemainingLength)
	_, err = io.ReadFull(r, packetBytes)
	if err != nil {
		return nil, err
	}
	err = cp.Unpack(bytes.NewBuffer(packetBytes))
	if err != nil {
		return nil, err
	}
	return cp, nil
}

func NewControlPacket(packetType byte) (cp ControlPacket) {
	switch packetType {
	case CONNECT:
		cp = &ConnectPacket{FixedHeader: FixedHeader{MessageType: CONNECT}}
	case CONNACK:
		cp = &ConnackPacket{FixedHeader: FixedHeader{MessageType: CONNACK}}
	case DISCONNECT:
		cp = &DisconnectPacket{FixedHeader: FixedHeader{MessageType: DISCONNECT}}
	case PUBLISH:
		cp = &PublishPacket{FixedHeader: FixedHeader{MessageType: PUBLISH}}
	case PUBACK:
		cp = &PubackPacket{FixedHeader: FixedHeader{MessageType: PUBACK}}
	case PUBREC:
		cp = &PubrecPacket{FixedHeader: FixedHeader{MessageType: PUBREC}}
	case PUBREL:
		cp = &PubrelPacket{FixedHeader: FixedHeader{MessageType: PUBREL, Qos: 1}}
	case PUBCOMP:
		cp = &PubcompPacket{FixedHeader: FixedHeader{MessageType: PUBCOMP}}
	case SUBSCRIBE:
		cp = &SubscribePacket{FixedHeader: FixedHeader{MessageType: SUBSCRIBE, Qos: 1}}
	case SUBACK:
		cp = &SubackPacket{FixedHeader: FixedHeader{MessageType: SUBACK}}
	case UNSUBSCRIBE:
		cp = &UnsubscribePacket{FixedHeader: FixedHeader{MessageType: UNSUBSCRIBE}}
	case UNSUBACK:
		cp = &UnsubackPacket{FixedHeader: FixedHeader{MessageType: UNSUBACK}}
	case PINGREQ:
		cp = &PingreqPacket{FixedHeader: FixedHeader{MessageType: PINGREQ}}
	case PINGRESP:
		cp = &PingrespPacket{FixedHeader: FixedHeader{MessageType: PINGRESP}}
	default:
		return nil
	}
	return cp
}

func NewControlPacketWithHeader(fh FixedHeader) (cp ControlPacket) {
	switch fh.MessageType {
	case CONNECT:
		cp = &ConnectPacket{FixedHeader: fh}
	case CONNACK:
		cp = &ConnackPacket{FixedHeader: fh}
	case DISCONNECT:
		cp = &DisconnectPacket{FixedHeader: fh}
	case PUBLISH:
		cp = &PublishPacket{FixedHeader: fh}
	case PUBACK:
		cp = &PubackPacket{FixedHeader: fh}
	case PUBREC:
		cp = &PubrecPacket{FixedHeader: fh}
	case PUBREL:
		cp = &PubrelPacket{FixedHeader: fh}
	case PUBCOMP:
		cp = &PubcompPacket{FixedHeader: fh}
	case SUBSCRIBE:
		cp = &SubscribePacket{FixedHeader: fh}
	case SUBACK:
		cp = &SubackPacket{FixedHeader: fh}
	case UNSUBSCRIBE:
		cp = &UnsubscribePacket{FixedHeader: fh}
	case UNSUBACK:
		cp = &UnsubackPacket{FixedHeader: fh}
	case PINGREQ:
		cp = &PingreqPacket{FixedHeader: fh}
	case PINGRESP:
		cp = &PingrespPacket{FixedHeader: fh}
	default:
		return nil
	}
	return cp
}

type Details struct {
	Qos       byte
	MessageID uint16
}

type FixedHeader struct {
	MessageType     byte
	Dup             bool
	Qos             byte
	Retain          bool
	RemainingLength int
}

func (fh FixedHeader) String() string {
	return fmt.Sprintf("%s: dup: %t qos: %d retain: %t rLength: %d", PacketNames[fh.MessageType], fh.Dup, fh.Qos, fh.Retain, fh.RemainingLength)
}

func boolToByte(b bool) byte {
	switch b {
	case true:
		return 1
	default:
		return 0
	}
}

func (fh *FixedHeader) pack() bytes.Buffer {
	var header bytes.Buffer
	header.WriteByte(fh.MessageType<<4 | boolToByte(fh.Dup)<<3 | fh.Qos<<1 | boolToByte(fh.Retain))
	header.Write(encodeLength(fh.RemainingLength))
	return header
}

func (fh *FixedHeader) unpack(typeAndFlags byte, r io.Reader) {
	fh.MessageType = typeAndFlags >> 4
	fh.Dup = (typeAndFlags>>3)&0x01 > 0
	fh.Qos = (typeAndFlags >> 1) & 0x03
	fh.Retain = typeAndFlags&0x01 > 0
	fh.RemainingLength = decodeLength(r)
}

func decodeByte(b io.Reader) byte {
	num := make([]byte, 1)
	b.Read(num)
	return num[0]
}

func decodeUint16(b io.Reader) uint16 {
	num := make([]byte, 2)
	b.Read(num)
	return binary.BigEndian.Uint16(num)
}

func encodeUint16(num uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, num)
	return bytes
}

func encodeString(field string) []byte {
	fieldLength := make([]byte, 2)
	binary.BigEndian.PutUint16(fieldLength, uint16(len(field)))
	return append(fieldLength, []byte(field)...)
}

func decodeString(b io.Reader) string {
	fieldLength := decodeUint16(b)
	field := make([]byte, fieldLength)
	b.Read(field)
	return string(field)
}

func decodeBytes(b io.Reader) []byte {
	fieldLength := decodeUint16(b)
	field := make([]byte, fieldLength)
	b.Read(field)
	return field
}

func encodeBytes(field []byte) []byte {
	fieldLength := make([]byte, 2)
	binary.BigEndian.PutUint16(fieldLength, uint16(len(field)))
	return append(fieldLength, field...)
}

func encodeLength(length int) []byte {
	var encLength []byte
	for {
		digit := byte(length % 128)
		length /= 128
		if length > 0 {
			digit |= 0x80
		}
		encLength = append(encLength, digit)
		if length == 0 {
			break
		}
	}
	return encLength
}

func decodeLength(r io.Reader) int {
	var rLength uint32
	var multiplier uint32 = 0
	b := make([]byte, 1)
	for {
		//todo: 是这里导致循环?
		_, err := io.ReadFull(r, b)
		if err != nil {
			return -1
		}
		digit := b[0]
		rLength |= uint32(digit&127) << multiplier
		if (digit & 128) == 0 {
			break
		}
		multiplier += 7
	}
	return int(rLength)
}
