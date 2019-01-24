package packets

import (
	"fmt"
	"io"
)

//PINGREQ packet

type PingreqPacket struct {
	FixedHeader
}

func (pr *PingreqPacket) String() string {
	str := fmt.Sprintf("%s", pr.FixedHeader)
	return str
}

func (pr *PingreqPacket) Write(w io.Writer) error {
	packet := pr.FixedHeader.pack()
	_, err := packet.WriteTo(w)

	return err
}

func (pr *PingreqPacket) Unpack(b io.Reader) error {
	return nil
}

func (pr *PingreqPacket) Details() Details {
	return Details{Qos: 0, MessageID: 0}
}
