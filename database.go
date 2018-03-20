package monica

import (
	"time"

	"git.domob-inc.cn/domob_pad/monica.git/config"

	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// global redis pool
	RedisPool *redis.Pool
)

// init beego `orm` config
// monica use beego's orm defaultly
// however we set the `DefaultRowsLimit` to -1 here
// the function read config from `default::mysql` part of yaml file
func InitDb() {
	configerMap, _ := config.Map("mysql")
	if configerMap == nil {
		return
	}

	orm.DefaultRowsLimit = -1
	orm.RegisterDriver("mysql", orm.DRMySQL)
	var initOk bool
	for key, value := range configerMap {
		if key == "default" {
			initOk = true
		}

		valueMap := value.(map[string]interface{})

		if err := orm.RegisterDataBase(key, "mysql",
			valueMap["dsn"].(string),
			valueMap["maxIdle"].(int),
			valueMap["maxOpen"].(int),
		); err != nil {
			panic(err)
		}
	}

	if !initOk {
		panic("a database instance called default must be inited")
	}

	runMode, _ := config.String("runmode")
	if runMode == "dev" {
		orm.Debug = true
	}

}

// init redis
// monica use `redigo` as a redis driver
// this function read config from config file
func InitRedis() {
	redisConfig, _ := config.Map("redis")
	if redisConfig == nil {
		return
	}

	address := redisConfig["address"].(string)
	if address == "" {
		panic("redis config not declared: get blank address")
	}
	db := redisConfig["db"].(int)
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
