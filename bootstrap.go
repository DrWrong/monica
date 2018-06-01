// 项目启动时的一些初始化配置

package monica

import (
	"fmt"
	"log"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/DrWrong/monica/config"
	"github.com/DrWrong/monica/logger"
	"github.com/DrWrong/monica/thriftext"
	"github.com/DrWrong/monica/webserver"
)

var (
	bootStrapLogger     *logger.MonicaLogger
	beforeQuiteHandlers []func()
	beforeQuiteWait     []func()
	WebServer           *webserver.WebServer
)

func init() {
	bootStrapLogger = logger.GetLogger("/monica/bootstrap")
}

// 注册退出前的处理函数
// 后注册的先处理
func RegisterBeforeQuiteHandler(handlers ...func()) {
	beforeQuiteHandlers = append(handlers, beforeQuiteHandlers...)
}

func RegisterBeforeQuitWait(waits ...func()) {
	beforeQuiteWait = append(waits, beforeQuiteWait...)
}

// 注册thrfit线程池
func RegisterThriftPool(poolname string, clientFactory interface{}) {
	field := fmt.Sprintf("thriftpool::%s", poolname)

	hosts, _ := config.Strings(
		fmt.Sprintf("%s::hosts", field))
	if len(hosts) == 0 {
		panic(fmt.Sprintf("%s hosts cannot be empty", poolname))
	}

	framed, _ := config.Bool(
		fmt.Sprintf("%s::framed", field))

	maxIdle, _ := config.Int(
		fmt.Sprintf("%s::max_idle", field))

	maxRetry, _ := config.Int(
		fmt.Sprintf("%s::max_retry", field))

	maxActive, _ := config.Int(fmt.Sprintf("%s::max_active", field))
	wait, _ := config.Bool(fmt.Sprintf("%s::wait", field))

	thriftext.GlobalThriftPool[poolname] = &thriftext.Pool{
		ClientFactory: clientFactory,
		Framed:        framed,
		Host:          hosts,
		MaxIdle:       maxIdle,
		MaxRetry:      uint(maxRetry),
		MaxActive:     maxActive,
		Wait:          wait,
	}

}

// 起动一个webserver
func BootStrapWeb(postInitFunc func()) {
	serverConfig, err := config.Map("server")
	if err != nil {
		panic(err)
	}

	webServerConfig := &webserver.ServerConfig{
		Port:       serverConfig["serverport"].(int),
		ServerMode: serverConfig["servermode"].(string),
	}

	if prefix, ok := serverConfig["urlPrefix"]; ok {
		webServerConfig.URLPrefix = prefix.(string)
	}

	WebServer = webserver.New(webServerConfig)

	postInitFunc()

	WebServer.Run()
}

// BooststrapThrift 起动thrift server
func BootstrapThrift(processor thrift.TProcessor) {
	port, err := config.Int("serverport")
	if err != nil {
		panic(err)
	}
	BootstrapThriftWithPort(processor, port)

}

func BootstrapThriftWithPort(processor thrift.TProcessor, port int) {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	networkAddr := fmt.Sprintf(":%d", port)
	serverTransport, err := thrift.NewTServerSocket(networkAddr)
	if err != nil {
		panic(err)
	}
	thriftServer := thrift.NewTSimpleServer4(processor, serverTransport, transportFactory, protocolFactory)
	log.Println("INFO: thrift server in", networkAddr)
	if err := thriftServer.Serve(); err != nil {
		panic(err)
	}

}
