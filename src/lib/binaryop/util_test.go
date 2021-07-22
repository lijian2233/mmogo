package binaryop

import (
	"fmt"
	"testing"
)

func TestInt16ToByte(t *testing.T) {
	val := int16(0x1122)
	b := [2]byte{}
	Int16ToByte(val, &b)
	v := BytesToInt16(b)
	fmt.Println(fmt.Sprintf("%x", b), fmt.Sprintf("%x",v))
}

func TestInt32ToBytes(t *testing.T) {
	val := int32(0x00123456)
	b := [4]byte{}
	Uint32ToBytes(uint32(val), &b)
	v := BytesToInt32(b)
	fmt.Println(fmt.Sprintf("%x", b), fmt.Sprintf("%x", v))
}

func TestInt64ToBytes(t *testing.T) {
	val := int64(0x1122334455667788)
	b := [8]byte{}
	Int64ToBytes(val, &b)
	v := BytesToInt64(b)
	fmt.Println(fmt.Sprintf("%x", b), fmt.Sprintf("%x", v))
}

func TestInt32ToByte(t *testing.T) {
	var n int64 = 0x1234567812345678
	fmt.Println(ebIntToByte(n))

	fmt.Println(ebIntToByte(SwapInt64(n)))
}