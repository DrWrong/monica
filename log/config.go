package log

import (
	"fmt"

	"github.com/DrWrong/monica/config"
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
	Name string
	Type string
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
	Name         string
	HandlerNames []string
	Level        Level
	Propagte     bool
}

func (option *LoggerOption) InitLogger() {
	handlers := make([]Handler, 0, len(option.HandlerNames))
	for _, handlerName := range option.HandlerNames {
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
		Propagte:   option.Propagte,
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

func InitLoggerFromConfigure(configure config.Configer) error {
	defer func() {
		initialized = true
	}()
	handlersConfig, err := configure.Maps("log::handlers")
	if err != nil {
		return err
	}
	handlerOptions := make([]*HandlerOption, 0, len(handlersConfig))
	for _, config := range handlersConfig {
		args := config["args"].(map[string]interface{})
		handlerOptions = append(handlerOptions, &HandlerOption{
			Name: config["name"].(string),
			Type: config["type"].(string),
			Args: args,
		})
	}

	loggerConfig, err := configure.Maps("log::loggers")
	if err != nil {
		return err
	}
	loggerOptions := make([]*LoggerOption, 0, len(loggerConfig))
	for _, config := range loggerConfig {
		level, _ := ParseLevel(config["level"].(string))
		handlers := config["handlers"].([]interface{})
		handlerNames := make([]string, 0, len(handlers))
		for _, handler := range handlers {
			handlerNames = append(handlerNames, handler.(string))
		}
		handlerNames = append(handlerNames)
		loggerOptions = append(loggerOptions, &LoggerOption{
			Name:         config["name"].(string),
			HandlerNames: handlerNames,
			Level:        level,
			Propagte:     config["propagte"].(bool),
		})
	}
	InitLogger(handlerOptions, loggerOptions)
	return nil
}

func ConfigFromFile(filename string) {
	configure := config.NewYamlConfig(filename)
	InitLoggerFromConfigure(configure)
}
