package packets

import (
	"fmt"
	"io"
)

//PUBREC packet

type PubrecPacket struct {
	FixedHeader
	MessageID uint16
}

func (pr *PubrecPacket) String() string {
	str := fmt.Sprintf("%s\n", pr.FixedHeader)
	str += fmt.Sprintf("MessageID: %d", pr.MessageID)
	return str
}

func (pr *PubrecPacket) Write(w io.Writer) error {
	var err error
	pr.FixedHeader.RemainingLength = 2
	packet := pr.FixedHeader.pack()
	packet.Write(encodeUint16(pr.MessageID))
	_, err = packet.WriteTo(w)

	return err
}

func (pr *PubrecPacket) Unpack(b io.Reader) error {
	pr.MessageID = decodeUint16(b)
	return nil
}

func (pr *PubrecPacket) Details() Details {
	return Details{Qos: pr.Qos, MessageID: pr.MessageID}
}
