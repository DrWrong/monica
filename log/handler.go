package log

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
)

type Handler interface {
	Handle(record *Record) error
}

type Rotator interface {
	shouldRollover(record *Record) bool
	doRollover()
}

type BaseHandler struct {
	writer      io.Writer
	logTemplate *template.Template
}

func (handler *BaseHandler) Handle(record *Record) error {
	return errors.New("not implement")
}

type FileHandler struct {
	*BaseHandler
	baseFileName string
}

func (handler *FileHandler) open() (io.Writer, error) {
	return os.OpenFile(handler.baseFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

}

func NewFileHandler(baseFileName, formatter string) (*FileHandler, error) {
	logTemplate := template.Must(template.New("logTemplate").Parse(formatter))
	handler := &FileHandler{
		BaseHandler:  &BaseHandler{logTemplate: logTemplate},
		baseFileName: baseFileName,
	}
	writer, err := handler.open()
	if err != nil {
		return nil, err
	}
	handler.writer = writer
	return handler, nil
}

func (handler *FileHandler) Handle(record *Record) error {
	if handler.writer == nil {
		handler.writer, _ = handler.open()
	}
	result, err := record.Bytes(handler.logTemplate)
	if err != nil {
		return err
	}
	handler.writer.Write(result)
	return nil
}

type RotatingFileHandler struct {
	handler *FileHandler
	rotator Rotator
}

func (handler *RotatingFileHandler) Handle(record *Record) error {
	if handler.rotator.shouldRollover(record) {
		handler.rotator.doRollover()
	}
	return handler.handler.Handle(record)
}

func NewTimeRotatingFileHandler(baseFileName, formatter, when string, backupCount int) (*RotatingFileHandler, error) {
	fileHandler, err := NewFileHandler(baseFileName, formatter)
	if err != nil {
		return nil, err
	}
	rotator := NewTimeRotator(when, backupCount, fileHandler)
	return &RotatingFileHandler{
		handler: fileHandler,
		rotator: rotator,
	}, nil
}

type TimeRotator struct {
	When        string
	BackupCount int
	interval    int64
	suffix      string
	extMatch    string
	rolloverAt  int64
	*FileHandler
}

func NewTimeRotator(when string, backupCount int, fileHandler *FileHandler) *TimeRotator {
	rotator := &TimeRotator{
		When:        strings.ToUpper(when),
		BackupCount: backupCount,
		FileHandler: fileHandler,
	}
	switch strings.ToUpper(when) {
	case "S":
		rotator.interval = 1
		rotator.suffix = "2006-01-02_15-04-05"
		rotator.extMatch = `^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}(\.\w+)?$`
	case "M":
		rotator.interval = 60
		rotator.suffix = "2006-01-02_15-04"
		rotator.extMatch = `^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}(\.\w+)?$`
	case "H":
		rotator.interval = 60 * 60
		rotator.suffix = "2006-01-02_15"
		rotator.extMatch = `^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}(\.\w+)?$`
	case "D":
		fallthrough
	case "MIDNIGHT":
		rotator.interval = 60 * 60 * 24
		rotator.suffix = "2006-01-02"
		rotator.extMatch = `^\d{4}-\d{2}-\d{2}(\.\w+)?$`
	default:
		panic("not validate rotator")
	}

	info, err := os.Stat(rotator.FileHandler.baseFileName)
	var currentTime int64
	if err != nil {
		currentTime = time.Now().Unix()
	} else {
		currentTime = info.ModTime().Unix()
	}
	rotator.rolloverAt = rotator.computeRollover(currentTime)
	return rotator
}

func (rotator *TimeRotator) computeRollover(currentTime int64) int64 {
	// case when rotator at middle night
	if rotator.When == "MIDNIGHT" {
		t := time.Unix(currentTime, 0)
		r := 24*60*60 - (t.Hour()*60+t.Minute())*60 + t.Second()
		return currentTime + int64(r)
	}
	return currentTime + rotator.interval
}

func (rotator *TimeRotator) shouldRollover(record *Record) bool {
	t := time.Now().Unix()
	return t >= rotator.rolloverAt
}

func (rotator *TimeRotator) doRollover() {
	if rotator.writer != nil {
		file := rotator.writer.(*os.File)
		file.Close()
		rotator.writer = nil
	}
	t := time.Unix(rotator.rolloverAt-rotator.interval, 0)
	dfn := rotator.baseFileName + "." + t.Format(rotator.suffix)
	os.Remove(dfn)
	os.Rename(rotator.baseFileName, dfn)
	if rotator.BackupCount > 0 {
		for _, fileName := range rotator.getFilesToDelete() {
			os.Remove(fileName)
		}
	}
	rotator.writer, _ = rotator.open()
	currentTime := time.Now().Unix()
	newRolloverAt := rotator.computeRollover(currentTime)
	for {
		if newRolloverAt > currentTime {
			break
		}
		newRolloverAt += rotator.interval
	}
	rotator.rolloverAt = newRolloverAt
}

func (rotator *TimeRotator) getFilesToDelete() []string {
	dirName, baseName := filepath.Split(rotator.baseFileName)
	fileInfos, _ := ioutil.ReadDir(dirName)
	result := make([]string, 0, 0)
	prefix := baseName + "."
	plen := len(prefix)
	for _, fileInfo := range fileInfos {
		name := fileInfo.Name()
		if len(name) <= plen {
			continue
		}

		if name[:plen] == prefix {
			suffix := name[plen:]
			if ok, _ :=regexp.MatchString(rotator.extMatch, suffix); ok {
				result = append(result, filepath.Join(dirName, name))
			}
		}
	}
	sort.Strings(result)
	if len(result) < rotator.BackupCount {
		result = make([]string, 0, 0)
	} else {
		result = result[:len(result)-rotator.BackupCount]
	}
	return result
}
