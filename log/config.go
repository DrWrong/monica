package log

import (
	"strconv"
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
		formatter := option.Args["formatter"]
		handler, _ = NewFileHandler(baseFileName, formatter)
	case "TimeRotatingFileHandler":
		baseFileName := option.Args["baseFileName"]
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


func init() {
	handlersMap = make(map[string]Handler, 0)
}
