DM303
=====
go版的dm303，在启动Go Server时启动dm303服务，则在Go Server通过dm303_go/stats可以进行简单的计数，可以通过thrift client来访问
计数的结果，来观察服务的状态。

在项目的src目录中:
```
	git clone http://git.domob-inc.cn/domoblib/dm303_go.git
```
启动Go Server的时候启动dm303即可:
```
package main

import "dm303_go"

func main() {
	go dm303_go.Start(10920, "just test")
	XXXXXX
}
```
Server中计数方式:
```
import "dm303_go/stats"

func XXXX() {
	stats.IncValue("asdasd")
	XXXX
	XXXX
	XXXX
}
```

版本号需要在程序的启动脚本中添加VERSION到环境变量中    
VERSION=1.1.3    
export VERSION    

Client端访问dm303服务时，注意要使用frame的方式
