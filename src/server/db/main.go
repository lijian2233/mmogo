package main

import (
	"flag"
	"fmt"
	"log"
	"mmogo/common"
	"mmogo/lib/packet"
	"mmogo/network/listen"
	"mmogo/server/db/global"
	"mmogo/server/db/start"
	"net"
	"runtime"
	"time"
)

var (
	configFile = flag.String("c", "config.toml", "config file")
	version    = flag.Bool("v", false, "version")
	buildTime  = "2018-01-01T00:00:00"
	gitHash    = "master"
)

func Start(confFile string)  {
	// set max cpu core
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := global.LoadFromFile(confFile); err != nil {
		log.Fatalf("parse config file error :%v", err)
	}

	if err := global.InitLogger(); err != nil {
		log.Fatalf("parse config file error :%v", err)
	}

	if err := global.InitGameDbMysql(); err != nil {
		log.Fatalf("init gorm open db error :%v", err)
	}

	if err := global.InitRedis(); err != nil {
		log.Fatalf("init redis error")
	}

	global.InitSocketMgr()

	start.WorkThreads.Start()

	listenSocket, err := listen.NewListenSocket(global.Conf.Server.Ip,
		global.Conf.Server.Port,
		listen.WithLog(global.Log),
		listen.WithHandler(start.GameAcceptConn))

	if err != nil {
		log.Fatalf("init redis error")
	}

	go listenSocket.Start()
	time.Sleep(time.Second * 1)
	exit := common.SetUpSignal()

	//go test()
	signal := <-exit
	global.Log.Infof("capture sigal %v", signal)

	global.Log.Infof("now close listen socket ..., wait at most 3 second")
	closeCh := listenSocket.Stop()
	select {
	case <-closeCh:
	case <-time.After(time.Second * 3):
	}

	global.Log.Infof("now close handle threads ..., wait at most 15 second")
	exitHandleThreadCh := make(chan bool, 1)
	go func() {
		start.WorkThreads.Stop()
		exitHandleThreadCh <- true
	}()

	select {
	case <-exitHandleThreadCh:
	case <-time.After(time.Second * 15):
	}

	global.Log.Infof("close db connect")
	global.GameDB.Close()

	global.Log.Infof("db stop complete")
}

func main() {
	// parse flag
	flag.Parse()

	if *version {
		fmt.Println("build time:", buildTime)
		fmt.Println("git hash:  ", gitHash)
		return
	}

	Start(*configFile)
}

func test()  {
	connStr := fmt.Sprintf("%s:%d", global.Conf.Server.Ip, global.Conf.Server.Port)
	conn, err := net.Dial("tcp", connStr)
	if err != nil {
		fmt.Println("test connect error ", err)
	}

	for i:=0; i<20; i++ {
		p := packet.NewWorldPacket(uint16(i), 0)
		p.SetSeqNo(uint32(i))
		n, err := conn.Write(p.GetPacket(false))
		fmt.Println(n, err)
	}
}