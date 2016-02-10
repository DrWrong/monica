package log

import (
	"testing"
	"time"
)

const formatter = `{{.Time.String }}  {{.Level.String }} {{.FileName }} {{.FuncName}} {{ .LineNo}} {{ .Message }}
`

func TestFileHanler(t *testing.T) {
	handler, err := NewFileHandler(
		"hanlder_test.log", formatter)
	if err != nil {
		t.Error(err)
	}
	record := NewRecord(DebugLevel, "this is a test")

	if err := handler.Handle(record); err != nil {
		t.Error(err)
	}
}

func TestTImeRotatingFileHandler(t *testing.T) {
	handler, err := NewTimeRotatingFileHandler("handler_test_time.log", formatter, "S", 10)
	if err != nil {
		t.Error(err)
	}

	record := NewRecord(DebugLevel, "this is a test")

	if err := handler.Handle(record); err != nil {
		t.Error(err)
	}
	time.Sleep(2 * time.Second)

	if err := handler.Handle(record); err != nil {
		t.Error(err)
	}

}
