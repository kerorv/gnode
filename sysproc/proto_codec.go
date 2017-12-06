package sysproc

type ProtoCodec struct {
}

func (pc *ProtoCodec) Encode(msg interface{}) (packet, bool) {
	return nil, false
}

func (pc *ProtoCodec) Decode(packet) (interface{}, bool) {
	return nil, false
}
