package common

import (
	"os"
	"os/signal"
	"syscall"
)

func SetUpSignal() <-chan os.Signal {
	// waitting for exit signal
	exit := make(chan os.Signal)
	stopSigs := []os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGABRT,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGUSR1,
	}
	signal.Notify(exit, stopSigs...)
	return exit
}
