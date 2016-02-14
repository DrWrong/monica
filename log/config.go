package log

import (
	"os"
	"path"
	"strconv"

	"github.com/DrWrong/monica/config"
)

var (
	handlersMap map[string]Handler
)

type HandlerOption struct {
	Name string
	Type string
	Args map[string]string
}

func (option *HandlerOption) InitHandler() {
	var handler Handler
	switch option.Type {
	case "FileHandler":
		baseFileName := option.Args["baseFileName"]
		baseFileName = path.Join(os.Getenv("MONICA_RUNDIR"), baseFileName)
		formatter := option.Args["formatter"]
		handler, _ = NewFileHandler(baseFileName, formatter)
	case "TimeRotatingFileHandler":
		baseFileName := option.Args["baseFileName"]
		baseFileName = path.Join(os.Getenv("MONICA_RUNDIR"), baseFileName)
		formatter := option.Args["formatter"]
		when := option.Args["when"]
		backupCount, _ := strconv.Atoi(option.Args["backupCount"])
		handler, _ = NewTimeRotatingFileHandler(baseFileName, formatter, when, backupCount)
	default:
		panic("not support handler type")
	}
	handlersMap[option.Name] = NewThreadSafeHandler(handler)
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
		handlers = append(handlers, handlersMap[handlerName])
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
}

func InitLoggerFromConfigure(configure config.Configer) {
	handlersConfig, err := configure.Maps("log::handlers")
	if err != nil {
		panic(err)
	}
	handlerOptions := make([]*HandlerOption, 0, len(handlersConfig))
	for _, config := range handlersConfig {
		args := config["args"].(map[string]interface{})
		argsConvert := make(map[string]string, len(args))
		for key, value := range args {
			argsConvert[key] = value.(string)
		}
		handlerOptions = append(handlerOptions, &HandlerOption{
			Name: config["name"].(string),
			Type: config["type"].(string),
			Args: argsConvert,
		})
	}

	loggerConfig, err := configure.Maps("log::loggers")
	if err != nil {
		panic(err)
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
}

func ConfigFromFile(filename string) {
	configure := config.NewYamlConfig(filename)
	InitLoggerFromConfigure(configure)
}

func init() {
	handlersMap = make(map[string]Handler, 0)
}
