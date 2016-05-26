package log

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Handler interface {
	Handle(record *Record) error
}

type ThreadSafeHandler struct {
	handler Handler
	sync.Mutex
}

func NewThreadSafeHandler(handler Handler) Handler {
	return &ThreadSafeHandler{
		handler: handler,
	}
}

func (handler *ThreadSafeHandler) Handle(record *Record) error {
	handler.Lock()
	defer handler.Unlock()
	return handler.handler.Handle(record)
}

type Rotator interface {
	shouldRollover(record *Record) bool
	doRollover()
}

type BaseHandler struct {
	writer      io.WriteCloser
	logTemplate *template.Template
}

func (handler *BaseHandler) Handle(record *Record) error {
	return errors.New("not implement")
}

// file handler process file
type FileHandler struct {
	*BaseHandler
	baseFileName string
}

// new a file handler
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

// provide a global method to new a file handler
func NewFileHandlerFactory(args map[string]interface{}) (Handler, error) {
	args1, ok := args["baseFileName"]
	if !ok {
		return nil, errors.New("baseFileName not exist in args")
	}
	baseFileName := path.Join(os.Getenv("MONIC_RUNDIR"), args1.(string))
	formatter, ok := args["formatter"]
	if !ok {
		return nil, errors.New("formatter not exist in args")
	}
	handler, err := NewFileHandler(baseFileName, formatter.(string))
	if err != nil {
		return nil, err
	}
	return NewThreadSafeHandler(handler), nil

}

func (handler *FileHandler) open() (io.WriteCloser, error) {
	return os.OpenFile(handler.baseFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

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

func NewTimeRotatingFileHandlerFactory(args map[string]interface{}) (Handler, error) {

	args1, ok := args["baseFileName"]
	if !ok {
		return nil, errors.New("baseFileName not exist in args")
	}
	baseFileName := path.Join(os.Getenv("MONIC_RUNDIR"), args1.(string))
	formatter, ok := args["formatter"]
	if !ok {
		return nil, errors.New("formatter not exist in args")
	}

	when, ok := args["when"]
	if !ok {
		return nil, errors.New("when not exist in args")
	}

	backupCount, ok := args["backupCount"]
	if !ok {
		return nil, errors.New("backupCount not exist in args")
	}

	handler, err := NewTimeRotatingFileHandler(
		baseFileName, formatter.(string), when.(string), backupCount.(int))

	if err != nil {
		return nil, err
	}

	return NewThreadSafeHandler(handler), nil

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
		// file := rotator.writer.(*os.File)
		// file.Close()
		rotator.writer.Close()
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
			if ok, _ := regexp.MatchString(rotator.extMatch, suffix); ok {
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

// send the log to a redis queue
type RedisHandler struct {
	Key         string
	logTemplate *template.Template
	pool        *redis.Pool
}

func (handler *RedisHandler) Handle(record *Record) error {
	conn := handler.pool.Get()
	defer conn.Close()
	result, err := record.Bytes(handler.logTemplate)
	if err != nil {
		return err
	}
	conn.Send("LPUSH", handler.Key, result)
	return conn.Flush()
}

// Redis handler factory
func NewRedisHandlerFactory(args map[string]interface{}) (Handler, error) {
	key, ok := args["key"]
	if !ok {
		return nil, errors.New("key not exist in args")
	}

	address, ok := args["address"]
	if !ok {
		return nil, errors.New("address not exist in args")
	}

	db, ok := args["db"]
	if !ok {
		return nil, errors.New("db not exist in args")
	}

	formatter, ok := args["formatter"]
	if !ok {
		return nil, errors.New("formatter not exit in args")
	}
	logTemplate := template.Must(template.New("logTemplate").Parse(formatter.(string)))

	pool := &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address.(string))
			if err != nil {
				println(err)
				return nil, err
			}
			_, err = c.Do("SELECT", db.(int))
			if err != nil {
				println(err)
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				println(err)
			}
			return err

		},
	}

	return &RedisHandler{
		Key:         key.(string),
		logTemplate: logTemplate,
		pool:        pool,
	}, nil
}


func init() {
	RegisterHandlerInitFunction("FileHandler", NewFileHandlerFactory)
	RegisterHandlerInitFunction("TimeRotatingFileHandler", NewTimeRotatingFileHandlerFactory)
	RegisterHandlerInitFunction("RedisHandler", NewRedisHandlerFactory)
}
