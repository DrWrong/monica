package monica

import (
	"time"

	"github.com/DrWrong/monica/log"
	"github.com/DrWrong/monica/config"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
)

var (
	RedisPool *redis.Pool
	dbLogger *log.MonicaLogger
)

func init() {
	dbLogger = log.GetLogger("/monica/database")
}

func InitDb() {
	orm.RegisterDriver("mysql", orm.DRMySQL)
	var initOk bool
	configerMap, _ := config.GlobalConfiger.Map("mysql")
	for key, value := range configerMap {
		if key == "default" {
			initOk = true
		}
		if err := orm.RegisterDataBase(key, "mysql", value.(string)); err != nil {
			panic(err)
		}
	}

	if !initOk {
		panic("a database instance called default must be inited")
	}

	if config.GlobalConfiger.String("runmode") == "dev" {
		orm.Debug = true
	}

}

func InitRedis() {
	address := config.GlobalConfiger.String("redis::address")
	if address == "" {
		panic("redis config not declared: get blank address")
	}
	db, _ := config.GlobalConfiger.Int("redis::db")
	RedisPool = &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				println(err)
				return nil, err
			}
			_, err = c.Do("SELECT", db)
			if err != nil {
				println(err)
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				println(err)
			}
			return err

		},
	}
}
