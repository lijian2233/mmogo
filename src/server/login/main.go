package login

import (
	"flag"
	"fmt"
	"runtime"
)

var (
	configFile = flag.String("c", "config.toml", "config file")
	version    = flag.Bool("v", false, "version")
	buildTime  = "2018-01-01T00:00:00"
	gitHash    = "master"
)

func main()  {
	// parse flag
	flag.Parse()

	if *version {
		fmt.Println("build time:", buildTime)
		fmt.Println("git hash:  ", gitHash)
		return
	}

	// set max cpu core
	runtime.GOMAXPROCS(runtime.NumCPU())



}

