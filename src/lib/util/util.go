package util

import (
	"fmt"
	"net"
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

func GetLocalIp() (ip string, err error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ip = ipnet.IP.String()
			return
		}
	}

	return
}

func GetInterfaceIp(name string) (ip string, err error) {
	inter, err := net.InterfaceByName(name)

	if err != nil {
		return
	}

	addrs, err := inter.Addrs()

	if err != nil {
		return
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ip = ipnet.IP.String()
			return
		}
	}

	return
}

