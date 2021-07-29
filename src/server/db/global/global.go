package global

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"mmogo/lib/logger"
	"mmogo/network/socket"
)

var Conf *config
var GameDB *gorm.DB

func load(file string) (*config, error) {
	c := config{}
	if _, err := toml.DecodeFile(file, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func LoadFromFile(file string) error {
	c, err := load(file)
	if err != nil {
		return err
	}
	Conf = c
	return nil
}


func InitGameDbMysql() error {
	if Conf == nil {
		panic("you must init config")
	}

	dbConf := Conf.GameDb
	connStr := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		dbConf.User, dbConf.Password, dbConf.Host, dbConf.Port, dbConf.Database, dbConf.Charset)

	db, err := gorm.Open("mysql", connStr)
	if err != nil {
		return err
	}

	if dbConf.MaxOpenConns < 20 {
		dbConf.MaxOpenConns = 20
	}
	db.DB().SetMaxOpenConns(dbConf.MaxOpenConns)

	if dbConf.MaxIdleConns < 5 {
		dbConf.MaxIdleConns = 5
	}
	db.DB().SetMaxIdleConns(dbConf.MaxIdleConns)

	db.LogMode(dbConf.Debug)

	GameDB = db
	return nil
}

var Log *logger.Logger

func InitLogger() error{
	if Conf == nil {
		panic("you must init config")
	}

	var err error
	Log, err = logger.NewLogger(&Conf.Logger)
	if err != nil {
		return err
	}
	return nil
}

var RedisClient *redis.Client
func InitRedis() error  {
	//连接服务器
	if Conf == nil {
		panic("you must init config")
	}
	redisdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", Conf.Redis.Host, Conf.Redis.Port), // use default Addr
		Password: Conf.Redis.Auth,               // no password set
		DB:       0,                // use default DB
	})

	_, err := redisdb.Ping().Result()
	if err != nil {
		return err
	}
	RedisClient = redisdb
	return nil
}

var SocketMgr *socket.GameSocketMgr

func InitSocketMgr()  {
	SocketMgr = socket.NewGameSocketMgr()
}