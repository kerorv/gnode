package sysproc

import (
	"encoding/binary"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/kerorv/gnode/tcp"
)

// ProtoCodec is a protobuf message codec
// protobuf message is len(2) + name(string, '\0' is end) + proto.Message([]byte)
type ProtoCodec struct {
}

func (pc *ProtoCodec) Encode(msg interface{}) (tcp.Packet, bool) {
	pb, ok := msg.(*proto.Message)
	if !ok {
		return nil, false
	}

	name := proto.MessageName(pb)
	nonProtoSize := tcp.PacketHeaderSize + len(name) + 1
	size := nonProtoSize + proto.Size(pb)
	p := tcp.MakePacket(size)
	p = append(p[tcp.PacketHeaderSize, []byte(name), 0)

	buf := proto.NewBuffer(p[nonProtoSize:])
	if err := buf.Marshal(pb); err != nil {
		return nil, false
	}

	return p, true
}

func (pc *ProtoCodec) Decode(p tcp.Packet) (interface{}, bool) {
	size := p.Size()
	var pos int
	for i, b := range p[tcp.PacketHeaderSize:] {
		if b == 0 {
			pos = tcp.PacketHeaderSize + i
			break
		}
	}

	if pos == 0 {
		return nil, false
	}

	name := string(p[tcp.PacketHeaderSize:pos])
	t := proto.MessageType(name)
	if t == nil {
		return nil, false
	}

	msg := reflect.New(t).Elem()
	return msg.Interface(), true
}
