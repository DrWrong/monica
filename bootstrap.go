// package monica provide some function like bootstrap or database initiation
// monica do some init functions when it is init
// firstly configure `GOMAXPROCS` to the `CPUNUMBER`
// sencodly it read some a global config file the file is finded in such a order(the `MONICA_CONFIGURE` environment variable, `config.yaml` in the current path, `$GOPATH/conf/monica.yaml`  )
// thridly it init the log system, register singal handler
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
	runDir := config.GlobalConfiger.String("default::runDir")
	if runDir != "" {
		os.Setenv("MONICA_RUNDIR", runDir)
	}
	// now config log module
	log.InitLoggerFromConfigure(config.GlobalConfiger)
	println("init logger ok")
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

// BootStrap a web server
// bootstrap firstly call the customizedConfig to do some init actions required by the upper programs
// then it start web server
func BootStrap(customizedConfig func()) {
	initGloabl(customizedConfig)
	// now start server
	server := core.NewServer()
	server.Run()
}

func initGloabl(customizedConfig func()) {
	bootStrapLogger.Debug("start init global config")
	if customizedConfig != nil {
		customizedConfig()
	}
	bootStrapLogger.Debug("init global config ok")

}

// BootStrap a thrift Server
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
