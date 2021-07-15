package conf

import (
	"github.com/BurntSushi/toml"
)

var Config Conf

type Conf struct {
	OpayChannelMysql MySQLConfig
	OpayStatisMysql  MySQLConfig
	Log              struct {
		Path       string //错误日志落地目录
		Level      string // debug  info  error
		TimeFormat string //time_format
		Color      bool
		Terminal   bool
		ShowIp     bool
	}

	Server struct {
		Port    string
		Mode    string
		GinMode string
	}
}

// MySQLConfig mysql 连接配置
type MySQLConfig struct {
	MaxOpenConns int    //最大连接数
	MaxIdleConns int    //最少启动连接数
	Master       string //主库
}

func ReadCfg(path string) Conf {
	if _, err := toml.DecodeFile(path, &Config); err != nil {
		panic("Parse File Error:" + err.Error())
	}
	return Config
}
