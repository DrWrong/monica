# Monica Logger

一个类python的日志管理系统, 与python的日志使用方法类似

## 使用方法

### 1. 进行日志配置

需要配置结构体 `LoggerConfig` , 并调用初始化函数

```golang

config := &LoggerConfig{
	Handlers: []*HandlerOption{
		&HandlerOption{
			// handler的名称
			Name: "fileHandler",
			// handler的类型
			Type: "Filehandler",
			// 初始化handler所需要的参数
			Args: map[string]interface{} {
				"baseFileName": "log_file.log",
				"formatter": "{{.Time.String }}  {{.Level.String }} {{.FileName }} {{.FuncName}} {{ .LineNo}} {{ .Message }} \n",
			}
		},
		&HandlerOption{
			// handler的名称
			Name: "timeRotatingHandler",
			// handler的类型
			Type: "TimeRotatingFileHandler",
			// 初始化handler所需要的参数
			Args: map[string]interface{} {
				"baseFileName": "time_rotating.log",
				"formatter": "{{.Time.String }}  {{.Level.String }} {{.FileName }} {{.FuncName}} {{ .LineNo}} {{ .Message }} \n",
				"when": "D",
				"backupCount": 10,
			}
		},

	},
	Loggers: []*LoggerOption{
		&LoggerOption{
			Name: "/monica/logger",
			// 上面配置的handler名称
			Handlers: []string{
				"fileHandler",
				"timeRotatingHandler",
			},
			Level: DebugLevel,
			Propagate: false,
		}
	}
}

// 调用初始化函数

InitLoggerByConfigure(config)


```

### 2. 获取logger实例

logger 实例的获取与 python稍有不同， 我们定义golang的logger以`/`来分割

形如 `/domob/ui` 其中 `/` 代表`rootLogger`, logger的查找形式也是如python那样层级化的， 如`/domob/ui` 首先会找名称为`/domob/ui`的 logger， 找不到时再去找`/domob`, 再找不到会去找`/` 这个`rootLogger`

另 `rootLogger`是默认就会配置的，默认情况下，`rootLogger` 输出到stderr的标准错误输出， 输出日志的格式为最复杂的形式


```golang

logger = GetLogger("/domob/ui")

```


### 3. 进行日志记录

目前定义了如下几个层级的日志 从高到低依次为

+ `Panic`
+ `Fatal`
+ `Error`
+ `Warn`
+ `Info`
+ `Debug`

进行日志记录时可选两种形式

1. 直接输出字符串 : `logger.Debug("this is a debug message")`
2. 格式化字符串： `logger.Errorf("process fail, error is: %+v", err)`


## 日志配置

### handler 配置

目前我们支持三种类型的hander: `FileHandler`, `TimeRotatingFileHandler`  和 `RedisHandler`

#### Formatter 参数

所有的Handler配置的`Args`中必须有一项是`formatter`, 用来约定如何输出日志。

formatter 实际上是采用了`text/template` 库, 所以配置文件的形式实际上是写了一个Template, Template传入的Record结构体如下：

```golang

type Record struct {
	// 日志的级别
	Level    Level
	// 日志的信息
	Message  string
	// 日志发生的时间
	Time     time.Time
	// 调用打日志的代码所在的文件名
	FileName string
	// 调用打日志的代码的行号
	LineNo   int
	// 调用打日志的代码所在的函数名称
	FuncName string
}

```

eg: 打印出所有信息: `"{{.Time.String }}  {{.Level.String }} {{.FileName }} {{.FuncName}} {{ .LineNo}} {{ .Message }} \n"`


#### FileHandler

+ 功能: 将日志输出到一个文件中不进行切分
+ 需要的参数

| 参数名称| 类型| 简介|
|----------|------|-------|
|`formatter`| string| 如上所述|
|`baseFileName`|string|输出到的文件|

**问题：**: 这里有一个问题， 我们在配置文件里一般会配置日志的相对路径， 然而这会存在一个问题，相对哪里？ 比如我们`go test`时，会`cd`到代码里面而不是模块根目录 这时配置相对路径就比较麻烦。 如果我们配置绝对路径，别人拿去代码又不能直接运行

**解决方案：** 我们引入了一个`MONICA_RUNDIR`的环境变量， 当设置了这个环境变量之后, 实际上输出到的文件为 `path.Join(os.Getenv("MONICA_RUNDIR", baseFileName))` 

#### TimeRotatingFileHandler

+ 功能：将日志输出到文件中并按照时间进行切分
+ 需要的参数

| 参数名称| 类型| 简介|
|----------|------|-------|
|`formatter`| string| 如上所述|
|`baseFileName`|string|输出到的文件|
|`when`|string|日志切分的时机: <br/> `S`每秒切割 <br/> `M` 每分切割 <br/> `H` 每小时切割 <br/> `D` 每天切割<br/> `MIDNIGHT` 每天整点晚上切割


#### RedisHandler

+ 功能: 使用`LPUSH` 将日志输出到redis的一个list中
+ 需要的参数

| 参数名称| 类型| 简介|
|----------|------|-------|
|`formatter`| string| 如上所述|
|`key`|string|输出到的key|
|`db`|int|使用的redis db|
|`address`|string|redis的地址`host:port` 的形式|


### logger 配置


logger的配置较为简单

```golang

type LoggerOption struct {
	// logger的名称 类 linux配置文件的模式 形如： "/" "/domob" "/domob/ui" 用"/"进行分级
	Name         string
	// logger所需要的处理的handler名称
	Handlers []string
	// logger所接受的level
	Level        Level
	// 是否向上反馈
	Propagate     bool
}


```
