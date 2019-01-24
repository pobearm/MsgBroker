package packets

import (
	"bytes"
	"fmt"
	"io"
)

//CONNACK packet

type ConnackPacket struct {
	FixedHeader
	TopicNameCompression byte
	ReturnCode           byte
}

func (ca *ConnackPacket) String() string {
	str := fmt.Sprintf("%s\n", ca.FixedHeader)
	str += fmt.Sprintf("returncode: %d", ca.ReturnCode)
	return str
}

func (ca *ConnackPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error

	body.WriteByte(ca.TopicNameCompression)
	body.WriteByte(ca.ReturnCode)
	ca.FixedHeader.RemainingLength = 2
	packet := ca.FixedHeader.pack()
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)

	return err
}

func (ca *ConnackPacket) Unpack(b io.Reader) error {
	ca.TopicNameCompression = decodeByte(b)
	ca.ReturnCode = decodeByte(b)
	return nil
}

func (ca *ConnackPacket) Details() Details {
	return Details{Qos: 0, MessageID: 0}
}
