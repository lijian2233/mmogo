package buffer

import (
	"fmt"
	"testing"
)

func TestRingBuffer(t *testing.T)  {
	//buf := make([]byte, 100, 100)
	ringBuffer := NewRingBuffer(100)

	w := ringBuffer.UnsafeWriteSpace()
	fmt.Println(&ringBuffer.buf[0], &w[0])

	fmt.Println(ringBuffer.Length(), ringBuffer.Free())
	for i:=0; i< 10; i++ {
		_, err := ringBuffer.WriteString("lijian111")
		if err != nil{
			fmt.Println(i, err)
		}
	}
	ringBuffer.Erase(20)
	w = ringBuffer.UnsafeWriteSpace()
	r := ringBuffer.UnsafeReadBytes()
	fmt.Println(&r[0], &w[0])

	fmt.Println(ringBuffer.Length(), ringBuffer.Free())
	ringBuffer.Erase(6)
	fmt.Println(ringBuffer.Length(), ringBuffer.Free())
}
