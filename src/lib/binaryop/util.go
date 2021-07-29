package binaryop

import (
	"bytes"
	eb "encoding/binary"
	"mmogo/lib/util"
	"unsafe"
)


func ebIntToByte(num interface{}) []byte {
	var buffer bytes.Buffer
	eb.Write(&buffer, eb.BigEndian, num)
	return buffer.Bytes()
}

func SwapInt16(n int16) int16 {
	return int16(SwapUint16(uint16(n)))
}

func SwapUint16(n uint16) uint16 {
	return n>>8 + (n&0x00FF)<<8
}

func SwapUint32(n uint32) uint32 {
	return n&0x000000FF<<24 + n&0xFF000000>>24 + n&0x00FF0000>>8 + n&0x0000FF00<<8
}

func SwapInt32(n int32) int32 {
	return int32(SwapUint32(uint32(n)))
}

func SwapUint64(n uint64) uint64 {
	a := uint32(n >> 32)
	b := uint32(n)
	return uint64(SwapUint32(a)) + uint64(SwapUint32(b))<<32
}

func SwapInt64(n int64) int64 {
	return int64(SwapUint64(uint64(n)))
}

func Int16ToByte(n int16, b *[2]byte) {
	if !util.IsLittleEnd() {
		b[0], b[1] = byte(n), byte(n>>8)
	} else {
		b[1], b[0] = byte(n), byte(n>>8)
	}
}

func Uint16ToByte(n uint16, b *[2]byte) {
	Int16ToByte(int16(n), b)
}

func Int32ToBytes(n int32, b *[4]byte) {
	if !util.IsLittleEnd() {
		b[0], b[1], b[2], b[3] = byte(n), byte(n>>8), byte(n>>16), byte(n>>24)
	} else {
		b[3], b[2], b[1], b[0] = byte(n), byte(n>>8), byte(n>>16), byte(n>>24)
	}
}

func Uint32ToBytes(n uint32, b *[4]byte) {
	Int32ToBytes(int32(n), b)
}

func Int64ToBytes(n int64, b *[8]byte) {
	if !util.IsLittleEnd() {
		b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7] = byte(n), byte(n>>8), byte(n>>16), byte(n>>24),
			byte(n>>32), byte(n>>40), byte(n>>48), byte(n>>56)
	} else {
		b[7], b[6], b[5], b[4], b[3], b[2], b[1], b[0] = byte(n), byte(n>>8), byte(n>>16), byte(n>>24),
			byte(n>>32), byte(n>>40), byte(n>>48), byte(n>>56)
	}
}

func Uint64ToBytes(n uint64, b *[8]byte) {
	Int64ToBytes(int64(n), b)
}

func BytesToInt16(b [2]byte) int16 {
	val := int16(0)
	if util.IsLittleEnd() {
		val = int16(b[0])<<8 + int16(b[1])
	} else {
		val = int16(b[1])<<8 + int16(b[0])
	}
	return val
}

func BytesToUint16(b [2]byte) uint16 {
	return uint16(BytesToInt16(b))
}

func BytesToInt32(b [4]byte) int32 {
	val := int32(0)
	if util.IsLittleEnd() {
		val = int32(b[0])<<24 + int32(b[1])<<16 + int32(b[2])<<8 + int32(b[3])
	} else {
		val = int32(b[3])<<24 + int32(b[2])<<16 + int32(b[1])<<8 + int32(b[0])
	}
	return val
}

func BytesToUint32(b [4]byte) uint32 {
	return uint32(BytesToInt32(b))
}

func BytesToInt64(b [8]byte) int64 {
	val := int64(0)
	if util.IsLittleEnd() {
		val = int64(b[0])<<56 + int64(b[1])<<48 + int64(b[2])<<40 + int64(b[3])<<32 +
			int64(b[4])<<24 + int64(b[5])<<16 + int64(b[6])<<8 + int64(b[7])

	} else {
		val = int64(b[7])<<56 + int64(b[6])<<48 + int64(b[5])<<40 + int64(b[4])<<32 +
			int64(b[3])<<24 + int64(b[2])<<16 + int64(b[1])<<8 + int64(b[0])
	}
	return val
}

func BytesToUInt64(b [8]byte) uint64 {
	return uint64(BytesToInt64(b))
}

func String2Byte(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	buf := *(*[]byte)(unsafe.Pointer(&h))
	return buf
}

func Bytes2Strings(bytes []byte) string {
	h := (*[3]uintptr)(unsafe.Pointer(&bytes))
	x := [2]uintptr{h[0], h[1]}
	str := *(*string)(unsafe.Pointer(&x))
	return str
}
