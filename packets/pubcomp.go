package packets

import (
	"fmt"
	"io"
)

//PUBCOMP packet

type PubcompPacket struct {
	FixedHeader
	MessageID uint16
}

func (pc *PubcompPacket) String() string {
	str := fmt.Sprintf("%s\n", pc.FixedHeader)
	str += fmt.Sprintf("MessageID: %d", pc.MessageID)
	return str
}

func (pc *PubcompPacket) Write(w io.Writer) error {
	var err error
	pc.FixedHeader.RemainingLength = 2
	packet := pc.FixedHeader.pack()
	packet.Write(encodeUint16(pc.MessageID))
	_, err = packet.WriteTo(w)

	return err
}

func (pc *PubcompPacket) Unpack(b io.Reader) error {
	pc.MessageID = decodeUint16(b)
	return nil
}

func (pc *PubcompPacket) Details() Details {
	return Details{Qos: pc.Qos, MessageID: pc.MessageID}
}
