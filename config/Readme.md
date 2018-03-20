# Monica config

一个配置文件管理的包

configer包含如下接口

```golang
// key的形式 a::b::c  用::表示Map的层级,
// 如果key不存在需要反回error
type Configer interface {
	String(key string) (string, error)
	Strings(key string) ([]string, error)
	Int(key string) (int, error)
	Ints(key string) ([]int, error)
	Float(key string) (float64, error)
	Floats(key string) ([]float64, error)
	Bool(key string) (bool, error)
	Bools(key string) ([]bool, error)
	Map(key string) (map[string]interface{}, error)
	Maps(key string) ([]map[string]interface{}, error)
}

```

##  使用方式

该模块的使用方式比较简单

1. 初始化配置文件， 目前只实现了`yaml` 的配置方式。之所以选择`yaml`是因为相对于ini yaml的说义更强 可以表达map， list 之类的， 相对于`json` yaml的容错性更强

调用`InitYamlGlobalConfiger(yamlFile)`来对`GlobalConfiger`实例进行初始化

2. 获取配置

获取配置直接调用`Configer`提供的接口即可。key是以`::`分割的来表示层级关系

比如配置文件

```yaml
server:
    serverport: 8205
    servermode: http
    urlPrefix: "/api"
```

调用方式

```golang

port, err := GlobalConfiger.Int("server::serverport")
mode, err := GlobalConfiger.String("server::servermode")
prefix, err := GlobalConfiger.String("server::urlPrefix")

```

为了简单起见， 我们提供了更简短的调用方式

```golang

port, err := Int("server::serverport")
mode, err := String("server::servermode")
prefix, err := String("server::urlPrefix")

```

## 实现

对于yaml的配置文件的实现，采用的方式是对第三方库`github.com/smallfish/simpleyaml`的简单封装
