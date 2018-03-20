package logger

import (
	"fmt"

)

var (
	handlersMap map[string]Handler = map[string]Handler{}

	handlerInitFunction map[string]FactoryFunc = map[string]FactoryFunc{}
)

func RegisterHandlerInitFunction(name string, initFunction FactoryFunc) {
	handlerInitFunction[name] = initFunction
}

// handler factory funct to create handler
type FactoryFunc func(map[string]interface{}) (Handler, error)

// the necessary options to init a handler
type HandlerOption struct {
	// handler 名称
	Name string
	// handler 类型
	Type string
	// 初始化handler所需要的参数
	Args map[string]interface{}
}

// a global init handler method
func (option *HandlerOption) InitHandler() {
	factoryFunc, ok := handlerInitFunction[option.Type]
	if !ok {
		panic("not support handler type")
	}

	handler, err := factoryFunc(option.Args)
	if err != nil {
		panic(err)
	}

	handlersMap[option.Name] = handler
}

type LoggerOption struct {
	// logger的名称 类 linux配置文件的模式 形如： "/" "/domob" "/domob/ui" 用"/"进行分级
	Name         string
	// logger所需要的处理的handler名称
	Handlers []string
	// logger所接受的level
	Level        Level
	// 是否向上反馈
	Propagate     bool
}

func (option *LoggerOption) InitLogger() {
	handlers := make([]Handler, 0, len(option.Handlers))
	for _, handlerName := range option.Handlers {
		handler, ok := handlersMap[handlerName]
		if !ok {
			panic(fmt.Sprintf("handler %s not exist", handlerName))
		}
		handlers = append(handlers, handler)
	}
	loggerMap[option.Name] = &MonicaLogger{
		handlers:   handlers,
		level:      option.Level,
		loggerName: option.Name,
		Propagate:   option.Propagate,
	}
}

func InitLogger(handlerOptions []*HandlerOption, loggerOption []*LoggerOption) {
	for _, option := range handlerOptions {
		option.InitHandler()
	}

	for _, option := range loggerOption {
		option.InitLogger()
	}
	initialized = true
}


type LoggerConfig struct {
	Handlers []*HandlerOption
	Loggers []*LoggerOption
}


// 初始化日志
func InitLoggerByConfigure(config *LoggerConfig) {
	defer func() {
		initialized = true
	}()

	InitLogger(config.Handlers, config.Loggers)
}


func PostInit() {
	initialized = true

}
