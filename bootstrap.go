package monica

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"
	"time"

	"github.com/DrWrong/monica/config"
	"github.com/DrWrong/monica/core"
	"github.com/DrWrong/monica/log"
	"github.com/DrWrong/monica/thrift"
)

var (
	bootStrapLogger *log.MonicaLogger
	thriftServer    *thrift.TSimpleServer
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// add default configure
	initGlobalConfiger()
	os.Setenv("MONICA_RUNDIR", config.GlobalConfiger.String("default::runDir"))
	// now config log module
	log.InitLoggerFromConfigure(config.GlobalConfiger)

	bootStrapLogger = log.GetLogger("/monica/bootstrap")
	go func() {
		for {
			c := make(chan os.Signal)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			//sig is blocked as c is 没缓冲
			sig := <-c
			bootStrapLogger.Infof("Signal %d received", sig)
			if thriftServer != nil {
				bootStrapLogger.Info("thrift server is going to stop")
				thriftServer.Stop()
				bootStrapLogger.Info("thrift server has gone away")
			}
			time.Sleep(time.Second)
			os.Exit(0)

		}
	}()
}

func BootStrap(customizedConfig func()) {
	initGloabl(customizedConfig)
	// now start server
	server := core.NewServer()
	server.Run()
}

func initGloabl(customizedConfig func()) {
	if customizedConfig != nil {
		customizedConfig()
	}

}

func BootStrapThriftServer(processor thrift.TProcessor, customizedConfig func()) {
	initGloabl(customizedConfig)

	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	port, _ := config.GlobalConfiger.Int("serverport")
	networkAddr := fmt.Sprintf(":%d", port)
	serverTransport, err := thrift.NewTServerSocket(networkAddr)
	if err != nil {
		bootStrapLogger.Errorf("Error! %s", err)
		os.Exit(1)
	}
	thriftServer = thrift.NewTSimpleServer4(processor, serverTransport, transportFactory, protocolFactory)
	fmt.Printf("thrift server in %s\n", networkAddr)
	if err := thriftServer.Serve(); err != nil {
		bootStrapLogger.Errorf("server start error: %s", err)
	}
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
