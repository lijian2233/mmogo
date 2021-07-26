package listen

import (
	"fmt"
	"net"
)

func defaultHandler() acceptFn{
	return handler
}

func handler(conn net.Conn)  {
	peer := conn.RemoteAddr()
	fmt.Println("remote %s conncet server", peer.String())
	fmt.Println("you should set up your handler, conn will be closed ...")
	conn.Close()
}