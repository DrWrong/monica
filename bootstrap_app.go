package monica

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/DrWrong/monica/config"
	"github.com/DrWrong/monica/logger"
	"gopkg.in/urfave/cli.v2"
)

const MONICA_PROCESS_CHILDREN_FLAG = "MONICA_PROCESS_CHILDREN_FLAG"

var App = NewMonicaApp()

func getConfigerFile() (configPath string) {
	// // check wheather filename from command line is empty
	// if configPath = c.String("config"); configPath != "" {
	//	return configPath
	// }

	dirName, _ := os.Getwd()
	// get config from current currentdir/config.yaml
	configPath = path.Join(dirName, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// get config path from currentdir/conf/monica.yaml
	configPath = path.Join(dirName, "conf", "monica.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// get config path from gopath
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		return ""
	}

	configPath = path.Join(goPath, "conf", "monica.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	return ""
}

// log的初始化
func initLogger() error {
	defer func() {
		logger.PostInit()
	}()
	handlersConfig, err := config.Maps("log::handlers")
	if err != nil {
		return err
	}
	handlerOptions := make([]*logger.HandlerOption, 0, len(handlersConfig))
	for _, config := range handlersConfig {
		args := config["args"].(map[string]interface{})
		handlerOptions = append(handlerOptions, &logger.HandlerOption{
			Name: config["name"].(string),
			Type: config["type"].(string),
			Args: args,
		})
	}

	loggerConfig, err := config.Maps("log::loggers")
	if err != nil {
		return err
	}
	loggerOptions := make([]*logger.LoggerOption, 0, len(loggerConfig))
	for _, config := range loggerConfig {
		level, _ := logger.ParseLevel(config["level"].(string))
		handlers := config["handlers"].([]interface{})
		handlerNames := make([]string, 0, len(handlers))
		for _, handler := range handlers {
			handlerNames = append(handlerNames, handler.(string))
		}
		handlerNames = append(handlerNames)
		loggerOptions = append(loggerOptions, &logger.LoggerOption{
			Name:      config["name"].(string),
			Handlers:  handlerNames,
			Level:     level,
			Propagate: config["propagte"].(bool),
		})
	}
	logger.InitLogger(handlerOptions, loggerOptions)
	return nil

}

func initGlobalConfig() bool {
	configPath := getConfigerFile()
	if configPath == "" {
		log.Println("WARNING: config path is empty so we will not load any config")
		return false
	}
	config.InitYamlGlobalConfiger(configPath)
	// now config log module
	if err := initLogger(); err != nil {
		// fmt.Fprintf(os.Stderr, "init from config error: load default configurre: %s", err)
		log.Println("WARNING: init logger from config error the logger will use default configure", err)
	}
	// init db from config
	InitDb()
	// init redis from config
	InitRedis()
	return true
}

type MonicaApp struct {
	*cli.App
	daemon             bool
	globalConfigInited bool
	name               string
}

func NewMonicaApp() *MonicaApp {
	app := &MonicaApp{
		globalConfigInited: initGlobalConfig(),
		name:               path.Base(os.Args[0]),
	}
	app.App = &cli.App{
		Name:  "Monica App",
		Usage: "Bootstrap an app quickly",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "daemon",
				Aliases:     []string{"d"},
				Usage:       "Specific that if you want to run process in daemon",
				Destination: &app.daemon,
			},
			&cli.StringFlag{
				Name:    "signal",
				Aliases: []string{"s"},
				Usage:   "send signal to process, accept args status, stop",
			},
		},
		Before: app.cliContextRedayHook,
	}
	return app
}

// 当初始化结束后执行的hook
func (app *MonicaApp) cliContextRedayHook(c *cli.Context) error {
	// 检查是否是子进程，如果不是将会开一个子进程并退出主进程
	if app.daemon {
		app.runInDaemonMode()
	}

	// check if signal specified
	signal := c.String("signal")
	if signal != "" {
		app.processSignal(signal)
		os.Exit(0)
	}
	// 能到此的进程是真正做事情的进程
	if app.daemon {
		app.writePidFile()
	}
	if app.globalConfigInited {
		StartDm303WhenNeccssary()
	}
	go app.handleSigIntAndTerm()
	return nil
}

func (app *MonicaApp) getPidFile() string {
	if app.globalConfigInited {
		pidFile, _ := config.String("pidfile")
		if pidFile != "" {
			// 绝对路径
			if strings.HasPrefix(pidFile, "/") {
				return pidFile
			}
			return path.Join(os.Getenv("MONICA_RUNDIR"), pidFile)
		}
	}
	runDir := os.Getenv("MONICA_RUNDIR")
	if runDir == "" {
		runDir, _ = os.Getwd()
	}
	return path.Join(runDir, "logs", fmt.Sprintf("%s.pid", app.name))
}

func (app *MonicaApp) getProcess() *os.Process {
	pidFile := app.getPidFile()
	rawPid, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return nil
	}

	pid, err := strconv.Atoi(string(rawPid))
	if err != nil {
		return nil
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	return p
}

func (app *MonicaApp) processAlive(p *os.Process) bool {
	if p == nil {
		p = app.getProcess()
	}

	if p == nil {
		return false
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		return false
	} else {
		return true
	}

}

func (app *MonicaApp) processSignal(signal string) {
	p := app.getProcess()
	if p == nil {
		log.Println("INFO: process is gone")
		return
	}
	// check status
	alive := app.processAlive(p)
	if !alive {
		log.Println("INFO: process is gone")
	} else if signal == "status" {
		log.Println("INFO: process is running at", p.Pid)
	}

	if alive && signal == "stop" {
		if err := p.Signal(syscall.SIGTERM); err != nil {
			log.Println("INFO: process is gone")
			return
		}

		for {
			time.Sleep(time.Second)
			if err := p.Signal(syscall.Signal(0)); err != nil {
				log.Println("INFO: process is gone")
				break
			}

			log.Println("INFO: process is still running wait it exit", p.Pid)
		}
	}
}

func (app *MonicaApp) writePidFile() {
	pidFile := app.getPidFile()
	f, err := os.OpenFile(pidFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	io.WriteString(f, fmt.Sprintf("%d", os.Getpid()))
}

func (app *MonicaApp) runInDaemonMode() {
	// is children then we do nothing
	if os.Getenv(MONICA_PROCESS_CHILDREN_FLAG) != "" {
		return
	}

	// first we should check if any file already running
	if app.processAlive(nil) {
		log.Fatalln("process is alive we will not start again...")
	}

	// spawn children process
	log.Println("INFO: starting process as daemon")
	_, err := os.StartProcess(
		os.Args[0],
		os.Args,
		&os.ProcAttr{
			Env: append(os.Environ(), fmt.Sprintf("%s=true", MONICA_PROCESS_CHILDREN_FLAG)),
			Files: []*os.File{
				nil,
				os.Stdout,
				os.Stderr,
			},
		},
	)
	if err != nil {
		log.Println("FATAL: create new process error")
		os.Exit(1)
	}
	time.Sleep(time.Second)
	if app.processAlive(nil) {
		log.Println("INFO: process is started")
	}
	os.Exit(0)
}

func (app *MonicaApp) handleSigIntAndTerm() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	//sig is blocked as c is 没缓冲
	sig := <-c
	log.Println("INFO: signal received", sig)
	log.Println("INFO: server is going to do some hourse keeping work")
	for _, handler := range beforeQuiteHandlers {
		handler()
	}
	log.Println("INFO: server is going to wait")
	for _, handler := range beforeQuiteWait {
		handler()
	}
	log.Println("INFO: byebye")
	if app.daemon {
		os.Remove(app.getPidFile())
	}
	os.Exit(0)
}
