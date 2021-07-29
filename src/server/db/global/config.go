package global

import (
	"mmogo/common"
	"mmogo/lib/logger"
)

type config struct {
	HandlerThreadNum uint8 `toml:"handler_thread_num"`
	GameDb           common.MySQLConfig
	Logger           logger.Config
	Redis            common.RedisConfig
	Server           common.ServerConfig
}
