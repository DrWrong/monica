package config

import (
	"sync"
)

var (
	GlobalConfiger     Configer
	globalServerOption *ServerOption
	once               sync.Once
)

func InitYamlGlobalConfiger(filename string) {
	GlobalConfiger = NewYamlConfig(filename)
}

type Configer interface {
	String(key string) string
	Strings(key string) []string
	Int(key string) (int, error)
	Ints(key string) ([]int, error)
	Float(key string) (float64, error)
	Floats(key string) ([]float64, error)
	Bool(key string) (bool, error)
	Bools(key string) ([]bool, error)
	Map(key string) (map[string]interface{}, error)
	Maps(key string) ([]map[string]interface{}, error)
}

// some server side options
type ServerOption struct {
	ServerPort int
	ServerMode string
	URLPrefix  string
}

// get a sigleton of gloabl configuration
// it will panic if global configer is null
func GetGlobalServerOption() *ServerOption {
	once.Do(initGloablServerOption)
	return globalServerOption
}

func initGloablServerOption() {
	globalServerOption = &ServerOption{
		ServerPort: 8200,
		ServerMode: "http",
	}
	if port, err := GlobalConfiger.Int("server::serverport"); err == nil {
		globalServerOption.ServerPort = port
	}

	if mode := GlobalConfiger.String("server::servermode"); mode == "http" || mode == "fastcgi" {
		globalServerOption.ServerMode = mode
	}

	globalServerOption.URLPrefix = GlobalConfiger.String("server::urlPrefix")

}
