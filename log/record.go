package log

import (
	"bytes"
	"fmt"
	"path"
	"runtime"
	"text/template"
	"time"
)

var srcFile string

func init() {
	_, file, _, _ := runtime.Caller(1)
	srcFile = path.Join(path.Dir(file), "logger.go")
}

type Level uint8

// Convert the Level to a string. E.g. PanicLevel becomes "panic".
func (level Level) String() string {
	switch level {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warning"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	}

	return "unknown"
}

// ParseLevel takes a string level and returns the Logrus log level constant.
func ParseLevel(lvl string) (Level, error) {
	switch lvl {
	case "panic":
		return PanicLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	}

	var l Level
	return l, fmt.Errorf("not a valid logrus Level: %q", lvl)
}

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `os.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
)

type Record struct {
	Level    Level
	Message  string
	Time     time.Time
	FileName string
	LineNo   int
	FuncName string
}

func NewRecord(level Level, message string) *Record {
	record := &Record{
		Level:   level,
		Message: message,
	}
	record.Time = time.Now()
	var pcs [2]uintptr
	numStack := runtime.Callers(4, pcs[:])
	for i := 0; i < numStack; i+=1 {
		function := runtime.FuncForPC(pcs[i])
		if function == nil {
			break
		}
		filename, line := function.FileLine(pcs[i])
		if filename == srcFile {
			continue
		} else {
			record.FileName = filename
			record.LineNo = line
			record.FuncName = function.Name()
			break
		}
	}

	return record
}

func (record *Record) Bytes(t *template.Template) (out []byte, err error) {
	var b bytes.Buffer
	err = t.Execute(&b, record)
	if err != nil {
		return
	}
	out = b.Bytes()
	return
}
