package binary

import (
	"mmogo/common"
)

func Int16ToByte(n int16, b *[2]byte){
	if !common.IsLittleEnd() {
		b[0], b[1] = byte(n), byte(n >> 8)
	} else {
		b[1], b[0] = byte(n), byte(n >> 8)
	}
}

func Uint16ToByte(n uint16, b *[2]byte)  {
	 Int16ToByte(int16(n), b)
}

func Int32ToBytes(n int32, b *[4]byte)  {
	if !common.IsLittleEnd() {
		b[0], b[1], b[2], b[3] = byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24)
	} else {
		b[3], b[2], b[1], b[0] = byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24)
	}
}

func Uint32ToBytes(n uint32, b *[4]byte) {
	 Int32ToBytes(int32(n), b)
}

func Int64ToBytes(n int64, b *[8]byte) {
	if !common.IsLittleEnd() {
		b[0],b[1],b[2],b[3],b[4],b[5],b[6],b[7] = byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24),
			byte(n >> 32), byte(n >> 40), byte(n >> 48), byte(n >> 56)
	} else {
		b[7],b[6],b[5],b[4],b[3],b[2],b[1],b[0] = byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24),
			byte(n >> 32), byte(n >> 40), byte(n >> 48), byte(n >> 56)
	}
}

func Uint64ToBytes(n uint64, b *[8]byte) {
	 Int64ToBytes(int64(n), b)
}

func BytesToInt16(b [2]byte) int16 {
	val := int16(0)
	if common.IsLittleEnd(){
		val = int16(b[0]) << 8 + int16(b[1])
	}else{
		val = int16(b[1]) << 8 + int16(b[0])
	}
	return val
}

func BytesToUint16(b [2]byte) uint16 {
	return uint16(BytesToInt16(b))
}

func BytesToInt32(b [4]byte) int32  {
	val := int32(0)
	if common.IsLittleEnd(){
		val = int32(b[0]) << 24 + int32(b[1]) << 16 + int32(b[2]) << 8 + int32(b[3])
	}else{
		val = int32(b[3]) << 24 + int32(b[2]) << 16 + int32(b[1]) << 8 + int32(b[0])
	}
	return val
}

func BytesToUint32(b [4]byte) uint32  {
	return uint32(BytesToInt32(b))
}

func BytesToInt64(b [8]byte) int64  {
	val := int64(0)
	if common.IsLittleEnd(){
		val = int64(b[0]) << 56 + int64(b[1]) << 48 + int64(b[2]) << 40 + int64(b[3]) << 32 +
			int64(b[4]) << 24 + int64(b[5]) << 16 + int64(b[6]) << 8 + int64(b[7])

	}else{
		val = int64(b[7]) << 56 + int64(b[6]) << 48 + int64(b[5]) << 40 + int64(b[4]) << 32 +
			int64(b[3]) << 24 + int64(b[2]) << 16 + int64(b[1]) << 8 + int64(b[0])
	}
	return val
}

func BytesToUInt64(b [8]byte) uint64  {
	return uint64(BytesToInt64(b))
}

