package tcp

type Codec interface {
	Encode(interface{}) (packet, bool)
	Decode(packet) (interface{}, bool)
}
