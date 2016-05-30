# monica
monica is a golang framework for the commonly server side situation.
This project is aimed to make a simple yet productive framework to ease the daily devlop task.


## What Monica Provide?

+ A Python like log system.
+ A yaml based config system.
+ A thrift extension to provide some functions like thrift client pool, json dump of thrift object.
+ A general used bootstrap function to start a server.
+ Some components like route, context, form binding to realize a web framework which provides REST service


## some use case


To realize a web server using `monica` you should do something like this:

```golang

import (
    "github.com/DrWrong/monica/core"
	"github.com/DrWrong/monica"
)

// in customer configure you can provide some config
// this function will be called after the system init finished
func customerConfigure() {
	fn := func(context *core.Context) {
		println("i am the handler")
		fmt.Printf("%+v\n", context)
     }
    // here we define some web root
    core.Handler(`^$`, fn)

}

// in the main function just call the bootStrap function with your config function

func main() {
    monica.BootStrap(customerConfigure)
}

```

Most commonly we need realize Thrift Server in this case Do Something like this

```golang

func main() {
    var (
       processor thrift.TProcessor
       cusomizedConfig func()
       )

    monica.BootStrapThriftServer(processor, customizedConfig)

}

```
