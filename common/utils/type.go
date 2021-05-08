package utils

import "encoding/binary"

// -------------------------------
// byte to uint
// -------------------------------

func ByteToUint16(buf []byte, littleEndian bool) uint16 {
	if littleEndian {
		return binary.LittleEndian.Uint16(buf[:2])
	} else {
		return binary.BigEndian.Uint16(buf[:2])
	}
}

func ByteToUint32(buf []byte, littleEndian bool) uint32 {
	if littleEndian {
		return binary.LittleEndian.Uint32(buf[:4])
	} else {
		return binary.BigEndian.Uint32(buf[:4])
	}
}

func ByteToUint64(buf []byte, littleEndian bool) uint64 {
	if littleEndian {
		return binary.LittleEndian.Uint64(buf[:8])
	} else {
		return binary.BigEndian.Uint64(buf[:8])
	}
}


// -------------------------------
// put uint to byte
// -------------------------------

func PutUint16ToByte(buf []byte, value uint16, littleEndian bool) {
	if littleEndian {
		binary.LittleEndian.PutUint16(buf[:2], value)
	} else {
		binary.BigEndian.PutUint16(buf[:2], value)
	}
}

func PutUint32ToByte(buf []byte, value uint32, littleEndian bool) {
	if littleEndian {
		binary.LittleEndian.PutUint32(buf[:4], value)
	} else {
		binary.BigEndian.PutUint32(buf[:4], value)
	}
}

func PutUint64ToByte(buf []byte, value uint64, littleEndian bool) {
	if littleEndian {
		binary.LittleEndian.PutUint64(buf[:8], value)
	} else {
		binary.BigEndian.PutUint64(buf[:8], value)
	}
}