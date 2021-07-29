package main

import (
	"flag"
	"fmt"
	"log"
	"mmogo/common"
	"mmogo/network/listen"
	"mmogo/server/db/global"
	"mmogo/server/db/handle"
	"runtime"
	"time"
)

var (
	configFile = flag.String("c", "config.toml", "config file")
	version    = flag.Bool("v", false, "version")
	buildTime  = "2018-01-01T00:00:00"
	gitHash    = "master"
)

func main() {
	// parse flag
	flag.Parse()

	if *version {
		fmt.Println("build time:", buildTime)
		fmt.Println("git hash:  ", gitHash)
		return
	}

	// set max cpu core
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := global.LoadFromFile(*configFile); err != nil {
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

	handle.HandleThreads.Start()

	listenSocket, err := listen.NewListenSocket(global.Conf.Server.Ip,
		global.Conf.Server.Port,
		listen.WithLog(global.Log),
		listen.WithHandler(handle.HandleAcceptConn))

	if err != nil {
		log.Fatalf("init redis error")
	}

	listenSocket.Start()
	exit := common.SetUpSignal()
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
		handle.HandleThreads.Stop()
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
