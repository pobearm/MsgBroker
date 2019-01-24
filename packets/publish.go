package packets

import (
	"bytes"
	"fmt"
	"io"
	"math"
)

//PUBLISH packet

type PublishPacket struct {
	FixedHeader
	TopicName string
	MessageID uint16
	Payload   []byte
}

func (p *PublishPacket) String() string {
	str := fmt.Sprintf("%s\n", p.FixedHeader)
	str += fmt.Sprintf("topicName: %s MessageID: %d\n", p.TopicName, p.MessageID)
	str += fmt.Sprintf("payload: %s\n", string(p.Payload))
	return str
}

func (p *PublishPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error

	body.Write(encodeString(p.TopicName))
	if p.Qos > 0 {
		body.Write(encodeUint16(p.MessageID))
	}
	p.FixedHeader.RemainingLength = body.Len() + len(p.Payload)
	packet := p.FixedHeader.pack()
	packet.Write(body.Bytes())
	packet.Write(p.Payload)
	_, err = packet.WriteTo(w)

	return err
}

// func (p *PublishPacket) SetRemainLength() {
// 	rl := 0
// 	if p.Qos > 0 {
// 		rl += len(p.TopicName) + 4
// 	} else {
// 		rl += len(p.TopicName) + 2
// 	}
// 	rl += len(p.Payload)
// 	p.FixedHeader.RemainingLength = rl
// }

func (p *PublishPacket) Unpack(b io.Reader) error {
	var payloadLength = p.FixedHeader.RemainingLength
	p.TopicName = decodeString(b)
	if p.Qos > 0 {
		p.MessageID = decodeUint16(b)
		payloadLength -= (len(p.TopicName) + 4)
	} else {
		payloadLength -= (len(p.TopicName) + 2)
	}
	if payloadLength < 0 || payloadLength > math.MaxInt32 {
		return fmt.Errorf("a error publishpackage unpack, payloadLength: %v, remainingLength: %v \n",
			payloadLength, p.FixedHeader.RemainingLength)

	}
	p.Payload = make([]byte, payloadLength)
	b.Read(p.Payload)
	return nil
}

func (p *PublishPacket) Copy() *PublishPacket {
	newP := NewControlPacket(PUBLISH).(*PublishPacket)
	newP.TopicName = p.TopicName
	newP.Payload = p.Payload

	return newP
}

func (p *PublishPacket) Details() Details {
	return Details{Qos: p.Qos, MessageID: p.MessageID}
}
