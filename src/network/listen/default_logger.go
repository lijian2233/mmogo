package listen

import (
	_interface "mmogo/interface"
	"mmogo/lib/logger"
)

func defualtLogger(name string) _interface.Logger {
	config := logger.DefaultConfig()
	config.Dir = config.Dir + "/" + name
	log, _ := logger.NewLogger(logger.DefaultConfig())
	return log
}