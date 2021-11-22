package utils

import "encoding/binary"

func Itob(i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
}
func Btoi(b []byte) int64 {
	switch len(b) {
	case 0:
		return 0
	case 1:
		return int64(b[0])
	case 2:
		return int64(binary.LittleEndian.Uint16(b))
	case 4:
		return int64(binary.LittleEndian.Uint32(b))
	case 8:
		return int64(binary.LittleEndian.Uint64(b))
	default:
		panic(`not support`)
	}
}
