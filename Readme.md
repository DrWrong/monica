# Monica golang framework

## 使用方式

`go get github.com/DrWrong/monica`

 
## 包含模块

+ [logger](logger) 一个类python的日志实现
+ [config](config) 一个配置文件的管理包
+ [webserver](webserver) 简单封装了一下`macaron` 用来扩展以支持**FastCGI**协议用于和旧的PHPUI进行无缝连接
+ [dm303_go](dm303_go) go 版本的dm303
+ [thrift](thrift) 将thrift代码拷贝了过来
+ [thriftext](thriftext) thriftext的扩展
+ [middleware](middleware) 中间件 用于web服务中的通用中间件


## Zen

尽量用环境变量与配置文件来控制程序的行为，以减少不必要的应用层代码

## Bootstrap逻辑

bootstrap帮你做了程序启动和退出时需要做的繁琐事情 这个的灵感来源于虫哥写的 java 的bootstrap

依次处理事情如下:

+ 设置golang的`MAXPROCS` 这在golang > 1.5之后是不必要的
+ 初始化日志配置 系统默认会依次从如下三个地方读取日志

> 1. 环境变量 `MONICA_CONFIGER`
> 2. 当前运行目录下的`config.yaml`
> 3. `GOPATH` 下面的 conf 下的 `monica.yaml`

+ 配置文件初始化 bootstrap会默认读取配置中的 `log`部分 并进行log配置。
+ 程序退出信号处理
+ 根据配置文件来配置mysql(db 用的是 beego的orm)
+ 根据配置文件来初始化redis (redis 用的是redigo)
+ 根据配置文件来启动dm303


## Bootstrap提供的一些便捷函数

+ `RegisterBeforeQuiteHandler(handlers ...func())` : 注入程序退出时的一系列处理函数 
+ `RegisterThriftPool(poolname string, clientFactory interface{})`: 注册thrift线程池
+ `BootStrapWeb(postInitFunc func())`: 启动webserver 自动从配置文件中读取web启动的相关配置
+ `BootStrapThrift(processor thrift.TProcessor)`: 启动thriftserver 自动从配置文件中读取thriftserver中的相关配置

## 配置文件式例


```yaml
# web server 配置
server:
    serverport: 8205
    servermode: http
    urlPrefix: "/api"

# mysql 配置
mysql:
  default: "domob:domob@tcp(dbm.office.domob-inc.cn:3306)/spring_promotion?charset=utf8mb4"

# redis 配置
redis:
  address: 10.0.0.207:16379
  db: 3

# dm303配置
dm303:
  service_name: "domob.dbb.ui"
  port: 8203

# thriftpool 配置
thriftpool:
  upserver:
    hosts:
      - "10.0.0.206:3091"
    fraxmed: true
    max_idle: 20
    max_retry: 2
    with_common_header: false


# log 的配置
log:
  handlers:
    - name: yamlfileHandler
      type: FileHandler
      args:
        baseFileName: "log/log_file.log"
        formatter: "{{.Time.String }} {{.FuncName }} {{.LineNo}} {{ .Message }} \n"
    - name: contextHandler
      type: TimeRotatingFileHandler
      args:
        baseFileName: "log/contxt.log"
        formatter: "{{.Time.String }}  {{ .Message }} \n"
        when: "D"
        backupCount: 2

  loggers:
    - name: /
      handlers:
        - yamlfileHandler
      level: debug
      propagte: false
    - name: /monica/core/context
      handlers:
        - contextHandler
      level: debug
      propagte: false

```






