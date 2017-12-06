package tcp

import (
	"encoding/binary"
	"math"
)

const (
	PacketHeaderSize = 2
	MaxPacketSize    = math.MaxUint16
)

type Packet []byte

func (p Packet) Size() uint16 {
	return binary.LittleEndian.Uint16(p)
}

func (p Packet) SetSize(size uint16) {
	binary.LittleEndian.PutUint16(p, size)
}

func MakePacket(size uint16) Packet {
	p := make(Packet, PacketHeaderSize, size)
	p.SetSize(size)
	return p
}
