package logger

import (
	"testing"
	"time"
)


func TestFileHanler(t *testing.T) {
	handler, err := NewFileHandler(
		"/dev/stdout", formatter)
	if err != nil {
		t.Error(err)
	}
	record := NewRecord(DebugLevel, "this is a test")

	if err := handler.Handle(record); err != nil {
		t.Error(err)
	}
}

func TestTimeRotatingFileHandler(t *testing.T) {
	handler, err := NewTimeRotatingFileHandler("handler_test_time.log", formatter, "S", 1)
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
