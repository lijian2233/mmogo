package common

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

var littleEnd bool = false

func IsLittleEnd() bool {
	return littleEnd
}

func testBigLittleEnd() {
	x := int16(0x1122)
	if byte(x) == byte(0x22) {
		fmt.Println("本机器：小端")
		littleEnd = true
	} else {
		fmt.Println("本机器：大端")
	}
}

func init() {
	testBigLittleEnd()
}

func GetGoid() int64 {
	var (
		buf [64]byte
		n   = runtime.Stack(buf[:], false)
		stk = strings.TrimPrefix(string(buf[:n]), "goroutine ")
	)

	idField := strings.Fields(stk)[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Errorf("can not get goroutine id: %v", err))
	}

	return int64(id)
}

