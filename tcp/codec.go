package tcp

type Codec interface {
	Encode(interface{}) (Packet, bool)
	Decode(Packet) (interface{}, bool)
}
