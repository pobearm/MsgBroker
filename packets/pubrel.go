package packets

import (
	"fmt"
	"io"
)

//PUBREL packet

type PubrelPacket struct {
	FixedHeader
	MessageID uint16
}

func (pr *PubrelPacket) String() string {
	str := fmt.Sprintf("%s\n", pr.FixedHeader)
	str += fmt.Sprintf("MessageID: %d", pr.MessageID)
	return str
}

func (pr *PubrelPacket) Write(w io.Writer) error {
	var err error
	pr.FixedHeader.RemainingLength = 2
	packet := pr.FixedHeader.pack()
	packet.Write(encodeUint16(pr.MessageID))
	_, err = packet.WriteTo(w)

	return err
}

func (pr *PubrelPacket) Unpack(b io.Reader) error {
	pr.MessageID = decodeUint16(b)
	return nil
}

func (pr *PubrelPacket) Details() Details {
	return Details{Qos: pr.Qos, MessageID: pr.MessageID}
}
