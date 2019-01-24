package packets

import (
	"bytes"
	"fmt"
	"io"
)

//SUBACK packet

type SubackPacket struct {
	FixedHeader
	MessageID   uint16
	GrantedQoss []byte
}

func (sa *SubackPacket) String() string {
	str := fmt.Sprintf("%s\n", sa.FixedHeader)
	str += fmt.Sprintf("MessageID: %d", sa.MessageID)
	return str
}

func (sa *SubackPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(encodeUint16(sa.MessageID))
	body.Write(sa.GrantedQoss)
	sa.FixedHeader.RemainingLength = body.Len()
	packet := sa.FixedHeader.pack()
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)

	return err
}

func (sa *SubackPacket) Unpack(b io.Reader) error {
	var qosBuffer bytes.Buffer
	sa.MessageID = decodeUint16(b)
	qosBuffer.ReadFrom(b)
	sa.GrantedQoss = qosBuffer.Bytes()
	return nil
}

func (sa *SubackPacket) Details() Details {
	return Details{Qos: 0, MessageID: sa.MessageID}
}
