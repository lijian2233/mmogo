package listen

import (
	_interface "mmogo/interface"
	"mmogo/lib/logger"
)

func defualtLog(name string) _interface.Log {
	config := logger.DefaultConfig()
	config.Dir = config.Dir + "/" + name
	log, _ := logger.NewLogger(logger.DefaultConfig())
	return log
}