package common

type ServerConfig struct {
	Ip   string
	Port uint16
	Mode string
}

// MySQLConfig mysql 连接配置
type MySQLConfig struct {
	MaxOpenConns int `toml:"max_open_conns"`//最大连接数
	MaxIdleConns int `toml:"max_idle_conns"`//最少启动连接数
	Host         string
	Port         int16
	User         string
	Password     string
	Charset      string
	Database     string
	Debug        bool
}

type RedisConfig struct {
	Host string
	Port int16
	Auth string
}

