package tcp

import (
	"encoding/binary"
	"math"
)

const (
	packetHeaderSize = 2
	maxPacketSize    = math.MaxUint16
)

type packet []byte

func (p packet) size() int {
	return int(binary.LittleEndian.Uint16(p))
}
