package test

import (
	"fmt"
	"mmogo/buffer"
	"testing"
	"unsafe"
)

const N int = int(unsafe.Sizeof(0))

func TestRingBuffer(t *testing.T)  {
	x := 0x1234
	p := unsafe.Pointer(&x)
	p2 := (*[N]byte)(p)
	if p2[0] == 0 {
		fmt.Println("本机器：大端")
	} else {
		fmt.Println("本机器：小端")
	}

	ringBuffer := buffer.NewRingBuffer(64*1024)
	ringBuffer.WriteString("lijian")
}

