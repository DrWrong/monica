### thriftext
thrift扩展

##  thrift client pool

本模块实现了一个golang 的 thrift client pool. 该模块参考redigo 的 pool来实现， 使用方法与redigo 的pool也基本一致

### 调用方法：

+ 初始化pool
```golang
    pool := NewThrfitPool{
            //创建client所需要的factory函数
            ClientFactory:aow_userserver.NewAowUserServerClientFactory,
            //可用的主机
            "10.0.0.206:38893",
            // 是否为framed
            true,
            // 最大空闲数
            10,
            // 最大重试数
            2
    }
```
+  获取一个client `client, _ := pool.Get()` 这里永远会获取到一个client 不必去关心是否有异常
+  调用某个方法  `result_as_interface, err := client.CallWithRetry(methodname, argumentlist ...)` 调用的第一个返回是interface类型的需要在业务层查询成所需要的类型, err 为application层异常，若不为application层的异常会尝试重试
+  关闭一个client  `client.Close()`

###  完整demo

```golang
	pool := &Pool{
		ClientFactory:    aow_userserver.NewAowUserServerClientFactory,
		Framed:           true,
		Host:             []string{"10.0.0.206:38895"},
		MaxIdle:          5,
		MaxRetry:         1,
		WithCommonHeader: true,
	}

	//获取client
	i := 1
	for {
		fmt.Printf("called %d times\n", i)
		i += 1
		device_info := &aow_userserver.DeviceInfo{
			Imei: "357143041410728",
		}

		// 调用方法1 适用于获取一个client 调用它的一个方法
		// 此处做了一层封装
		fmt.Print("call with method 1st")
		id, err := pool.CallWithRetry("GetUserId", device_info, true)
		fmt.Println(id)
		if err != nil {
			fmt.Println(err)
		}
		// 调用方法2 使用于获取一个client并调用它的一系列方法然后返回
		fmt.Print("called with method 2nd")
		client, _ := pool.Get()
		id, err = pool.CallWithRetry("GetUserId", device_info, true)
		fmt.Println(id)
		if err != nil {
			fmt.Println(err)
		}
		// 可以调用一些其他的方法然后关掉
		// 在正式开发中应该是用defer
		client.Close()
		time.Sleep(5 * time.Second)
	}
}
```

### call with retry
 为方便使用，我们封装了 CallWithRetry 方法。
初始化pool
```golang
    pool := &Pool{
        ClientFactory: aow_userserver.NewAowUserServerClientFactory,
        Framed:        true,
        Host:          []string{"10.0.0.206:38895"},
        MaxIdle:       5,
        MaxRetry:      5,
    }
```
Pool级别的retry, 此方法适用于获取一个client并调用单个的方法
```golang
resinterface, err := pool.CallWithRetry(methodname, args...)
```

Client 级别的retry, 此方法适用于获取一个client，并调用它的多个方法

```golang
client, err := pool.Get()
if err != nil {
    fmt.Println(err)
}
if client == nil {
    break
}
client.CallWithRetry(methodname, args...)
```



