package log

import (
	"testing"
)

func TestRootLogger(t *testing.T) {
	Debug("this is a test")
}

func TestGetLogger(t *testing.T) {
	handlerOptions := []*HandlerOption{
		&HandlerOption{
			Name: "fileHandler",
			Type: "FileHandler",
			Args: map[string]string{
				"baseFileName": "logger_test.log",
				"formatter":    formatter,
			},
		},
	}
	loggerOptions := []*LoggerOption{
		&LoggerOption{
			Name:         "/monica/logger",
			HandlerNames: []string{"fileHandler"},
			Level:        DebugLevel,
			Propagte:     false,
		},
	}
	InitLogger(handlerOptions, loggerOptions)
	logger := GetLogger("/monica/logger")
	t.Log(logger.loggerName)
	logger.Debug("test get logger")
}

func TestFileLogger(t *testing.T) {
	ConfigFromFile("log.yaml")
	logger := GetLogger("/monica/filelogger")
	t.Log(logger.loggerName)
	logger.Debug("test file logger")
}
