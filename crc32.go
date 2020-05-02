package window_crc32

const (
	CrcPoly = 0xEDB88320
)

type WindowCRC32 interface {
	UpdateCRC32(b byte)
	Finish() uint
}

func NewCRC(poly uint, window uint, s Serializer) WindowCRC32 {
	return newCRC(poly, window, s, 1)
}

func NewCRCThreaded(poly uint, window uint, s Serializer, threads uint) WindowCRC32 {
	return newCRC(poly, window, s, threads)
}

func NewJsonSerializer(mod uint) Serializer {
	return &jsonSerializer{memSerializer{mod: mod}}
}
