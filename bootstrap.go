package monica

import (
	"os"
	"path"

	"github.com/DrWrong/monica/config"
	"github.com/DrWrong/monica/core"
	"github.com/DrWrong/monica/log"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func BootStrap(customizedConfig func()) {
	// add default configure
	initGlobalConfiger()
	os.Setenv("MONICA_RUNDIR", config.GlobalConfiger.String("default::runDir"))
	// now config log module
	log.InitLoggerFromConfigure(config.GlobalConfiger)

	// now config db module we use beego as default db
	initDb()

	if customizedConfig != nil {
		customizedConfig()
	}
	// now start server
	server := core.NewServer()
	server.Run()
}

func initGlobalConfiger() {
	// firstly get configure file from environment config
	configPath := os.Getenv("MONICA_CONFIGER")
	if configPath != "" {
		config.InitYamlGlobalConfiger(configPath)
		return
	}
	// sencodly examine if a config file exist on the current dir
	dirName, _ := os.Getwd()
	configPath = path.Join(dirName, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		config.InitYamlGlobalConfiger(configPath)
		return
	}

	// thirdly examine if a config file exist on the GOPATH

	goPath := os.Getenv("GOPATH")
	configPath = path.Join(goPath, "conf", "monica.yaml")
	if _, err := os.Stat(configPath); err == nil {
		config.InitYamlGlobalConfiger(configPath)
		return
	}
	panic("no config file found!!!")
}

func initDb() {
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
