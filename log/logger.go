package log

import (
	"fmt"
	"path"
)

const formatter = `{{.Time.String }}  {{.Level.String }} {{.FileName }} {{.FuncName}} {{ .LineNo}} {{ .Message }}
`

// save loggers in a tree like structure
var (
	loggerMap          map[string]*MonicaLogger
	propagateLoggerMap map[string][]*MonicaLogger
	initialized        bool
)

func GetLogger(name string) *MonicaLogger {
	if !initialized {
		return &MonicaLogger{
			loggerPath: name,
			isCache:    true,
		}
	}
	for {
		logger, ok := loggerMap[name]
		if ok {
			return logger
		}
		if name == "/" && !ok {
			panic("root logger not configured")
		}
		name = path.Dir(name)
	}
}

func getRootLogger() *MonicaLogger {
	return GetLogger("/")
}

func getParentLoggers(name string) []*MonicaLogger {
	loggers := make([]*MonicaLogger, 0, 0)
	if name == "/" {
		return loggers
	}
	for {
		name = path.Dir(name)
		logger, ok := loggerMap[name]
		if ok {
			loggers = append(loggers, logger)
		}
		if name == "/" {
			break
		}

	}
	return loggers
}

// this function is not thread safe, but it do not have too much problems
func getParentLoggersCache(name string) []*MonicaLogger {
	if loggers, ok := propagateLoggerMap[name]; ok {
		return loggers
	}
	loggers := getParentLoggers(name)
	propagateLoggerMap[name] = loggers
	return loggers
}

func Debug(msg string) {
	getRootLogger().Debug(msg)
}

func Debugf(formatter string, args ...interface{}) {
	getRootLogger().Debugf(formatter, args...)
}

func Info(msg string) {
	getRootLogger().Info(msg)
}

func Infof(formatter string, args ...interface{}) {
	getRootLogger().Infof(formatter, args...)
}

func Warn(msg string) {
	getRootLogger().Warn(msg)
}

func Warnf(formatter string, args ...interface{}) {
	getRootLogger().Warnf(formatter, args...)
}

func Error(msg string) {
	getRootLogger().Error(msg)
}

func Errorf(formatter string, args ...interface{}) {
	getRootLogger().Errorf(formatter, args...)
}

func Fatal(msg string) {
	getRootLogger().Fatal(msg)
}

func Fatalf(formatter string, args ...interface{}) {
	getRootLogger().Fatalf(formatter, args...)
}

type MonicaLogger struct {
	handlers   []Handler
	level      Level
	loggerName string
	loggerPath string
	isCache    bool
	Propagte   bool
}

func (logger *MonicaLogger) logEmit(record *Record) {
	// if logger level is not satisfied just ignore the record
	if record.Level > logger.level {
		return
	}
	for _, handler := range logger.handlers {
		handler.Handle(record)
	}

}

func (logger *MonicaLogger) log(level Level, msg string) {
	if logger.isCache {
		if !initialized {
			panic("cannot use logger before initialize")
		}
		logger = GetLogger(logger.loggerPath)
	}
	record := NewRecord(level, msg)
	logger.logEmit(record)
	if logger.Propagte {
		for _, logger := range getParentLoggersCache(logger.loggerName) {
			logger.logEmit(record)

		}
	}
}

func (logger *MonicaLogger) Debug(msg string) {
	logger.log(DebugLevel, msg)
}

func (logger *MonicaLogger) Debugf(format string, args ...interface{}) {
	logger.log(DebugLevel, fmt.Sprintf(format, args...))
}

func (logger *MonicaLogger) Info(msg string) {
	logger.log(InfoLevel, msg)
}

func (logger *MonicaLogger) Infof(format string, args ...interface{}) {
	logger.log(InfoLevel, fmt.Sprintf(format, args...))
}

func (logger *MonicaLogger) Warn(msg string) {
	logger.log(WarnLevel, msg)
}

func (logger *MonicaLogger) Warnf(format string, args ...interface{}) {
	logger.log(WarnLevel, fmt.Sprintf(format, args...))
}

func (logger *MonicaLogger) Error(msg string) {
	logger.log(ErrorLevel, msg)
}

func (logger *MonicaLogger) Errorf(format string, args ...interface{}) {
	logger.log(ErrorLevel, fmt.Sprintf(format, args...))
}

func (logger *MonicaLogger) Fatal(msg string) {
	logger.log(FatalLevel, msg)
}

func (logger *MonicaLogger) Fatalf(format string, args ...interface{}) {
	logger.log(FatalLevel, fmt.Sprintf(format, args...))
}

func init() {
	propagateLoggerMap = make(map[string][]*MonicaLogger, 0)
	handler, _ := NewFileHandler("/dev/stdout", formatter)
	rootLogger := &MonicaLogger{
		handlers:   []Handler{NewThreadSafeHandler(handler)},
		level:      DebugLevel,
		loggerName: "/",
	}
	loggerMap = make(map[string]*MonicaLogger, 0)
	loggerMap["/"] = rootLogger
}
