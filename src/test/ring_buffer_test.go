package test

import (
	"fmt"
	"mmogo/buffer"
	"mmogo/common"
	"testing"
	"unsafe"
)

const N int = int(unsafe.Sizeof(0))

func TestRingBuffer(t *testing.T)  {

	str := "lijian"
	bytes := common.String2Byte(str)
	str1 := common.Bytes2Strings(bytes)
	fmt.Println(str1)
	x := 0x1234
	p := unsafe.Pointer(&x)
	p2 := (*[N]byte)(p)
	if p2[0] == 0 {
		fmt.Println("本机器：大端")
	} else {
		fmt.Println("本机器：小端")
	}

	ringBuffer := buffer.NewRingBuffer(10)
	ringBuffer.WriteString("lijian")
	str2 := common.Bytes2Strings(ringBuffer.UnsafeBytes())
	ringBuffer.Erase(len(str2))
	ringBuffer.WriteString("lijian")
	ringBuffer.Erase(5)
	fmt.Println(str2)
}

