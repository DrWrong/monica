package thriftext

// 本文件实现了一个 go 的 thrift 链接池， 具体使用可以参照测试用例
import (
	"container/list"
	"domob_thrift/common"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/DrWrong/monica/config"
	"github.com/DrWrong/monica/thrift"
)

var (
	GlobalThriftPool map[string]Pool
)

// 定义一个thrift client 的接口
type ThriftClient interface {
}

// 关闭连接
func closeTransport(client ThriftClient) {
	if client == nil {
		return
	}
	reflectClient := reflect.ValueOf(client).Elem()
	transport := reflectClient.FieldByName("Transport").Interface().(thrift.TTransport)
	transport.Close()
}

// //定义了一个封装的thrift cliet 该类有一个封装的Close方法用于将client返回到类中
type WrappedClient struct {
	p         *Pool
	client    ThriftClient
	err       error
	borrowNum int //被调用的次数
}

func (w *WrappedClient) Close() {
	w.p.put(w)
}

// 重试机制
func (w *WrappedClient) CallWithRetry(name string, args ...interface{}) (res interface{}, err error) {
	res, err = w.Call(name, args...)
	// 出错后进行重试，链接层的错误一般是服务重启引起的，而服务重启会导致所有的client全部不可用, 这是一个消极的方案。
	if w.err != nil {
		// 关闭旧有连接
		w.Close()

		var i uint
		for i = 0; i < w.p.MaxRetry; i += 1 {
			// 如果重试次数超过2次就武断的判定为所有的client都不可以用全都关掉
			if i == 1 {
				w.p.closeAllClient()
			}
			// sleep 一会
			t := i
			if t >= 5 {
				t = 5
			}
			time.Sleep((1 << t) * time.Second)

			// 获取一个新的连接并城市调用
			w, _ = w.p.Get()
			res, err = w.Call(name, args...)
			// 如果调用出错
			if w.err != nil {
				// 把新获取到的连接关掉
				fmt.Printf("retry %d times\n", i)
				w.Close()
				continue
			}
			return res, err
		}
	}
	return res, err
}

// client的方法调用
func (w *WrappedClient) Call(name string, args ...interface{}) (response interface{}, err error) {
	// 如果client本身有问题
	if w.err != nil {
		println(w.err.Error())
		return nil, errors.New("the client get have some errors")
	}
	client := reflect.ValueOf(w.client)
	method := client.MethodByName(name)
	values := make([]reflect.Value, 0, len(args)+1)
	if w.p.WithCommonHeader {
		header := common.NewRequestHeader()
		values = append(values, reflect.ValueOf(header))
	}
	for _, arg := range args {
		values = append(values, reflect.ValueOf(arg))
	}
	// 返回结果不确定，可能是1个或者两个
	res := method.Call(values)

	var errResponse interface{}

	if len(res) == 1 {
		errResponse = res[0].Interface()
	} else if len(res) == 2 {
		response = res[0].Interface()
		errResponse = res[1].Interface()
	} else {
		response = res[0].Interface()
		errResponse = res[len(res)-1].Interface()
	}
	if errResponse != nil {
		err = errResponse.(error)
		if _, ok := err.(thrift.TApplicationException); !ok {
			w.err = err
		}
	}
	return
}

//pool layer
type Pool struct {
	// 连接初始化函数
	ClientFactory interface{}
	// 是否为Framed
	Framed bool
	//  如主机列表 '10.0.0.206:8087' 形式
	Host []string
	// maximum number of idle connections in the pool
	MaxIdle int
	// 允许的最大链接
	MaxActive int
	//  是否阻塞
	Wait bool
	// 是否使用通用header
	WithCommonHeader bool
	MaxRetry         uint
	// mu protects fields defined below
	mu   sync.Mutex
	cond *sync.Cond
	// 现有链接
	active int
	// 存放空闲的client
	idle list.List
}

func NewThriftPool(clientFactory interface{}, host []string,
	framed bool, maxIdle int, maxRetry uint,
	withComonHeader bool) *Pool {
	return &Pool{
		ClientFactory:    clientFactory,
		Framed:           framed,
		Host:             host,
		MaxIdle:          maxIdle,
		MaxRetry:         maxRetry,
		WithCommonHeader: withCommonHeader, //默认下支持多个头部
	}
}

// release decrements the active count and signals waiters. The caller must
// hold p.mu during the call.
// 释放一个现有链接
func (p *Pool) release() {
	p.active -= 1
	if p.cond != nil {
		p.cond.Signal()
	}
}

// 新建一个client的具体操作
func (p *Pool) newThriftClient() (thriftClient ThriftClient, err error) {
	// transport 层可能有两种transport
	var transportFactory thrift.TTransportFactory
	if p.Framed {
		transportFactory = thrift.NewTTransportFactory()
		transportFactory = thrift.NewTFramedTransportFactory(transportFactory)
	} else {
		transportFactory = thrift.NewTBufferedTransportFactory(8192)
	}
	// protocol 层只用binary protocol
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	// 随机选取第一个可用的host，如果所有host都不可用那么抛出异常
	hostLen := len(p.Host)
	randHostSlect := rand.Perm(hostLen)
	// host := p.Host[hostLen/randNum]
	var transport thrift.TTransport
	for _, idx := range randHostSlect {
		host := p.Host[idx]
		transport, err = thrift.NewTSocket(host)
		if err != nil {
			continue
		}
		transport = transportFactory.GetTransport(transport)
		err = transport.Open()
		if err == nil {
			break
		}
	}
	// 如果所有主机的连接都打不开，那么抛异常
	if err != nil {
		return nil, err
	}

	clientFactory := reflect.ValueOf(p.ClientFactory)
	args := []reflect.Value{
		reflect.ValueOf(transport), reflect.ValueOf(protocolFactory)}
	res := clientFactory.Call(args)
	thriftClient = res[0].Interface()
	return thriftClient, nil

}

// 关掉所有资源时pool不可用
func (p *Pool) closeAllClient() {
	log.Println("now close all the client")
	p.mu.Lock()
	for {
		e := p.idle.Back()
		if e == nil {
			break
		}
		p.idle.Remove(e)
		// 将链接关掉并且释放
		client := e.Value.(*WrappedClient)
		closeTransport(client.client)
		p.release()
	}

	p.mu.Unlock()

}

// 获取一个client
func (p *Pool) Get() (*WrappedClient, error) {
	p.mu.Lock()
	for {
		// 从list中取出一个client来
		for {
			e := p.idle.Back()
			if e == nil {
				break
			}
			p.idle.Remove(e)
			// 维护一个引用计数
			client := e.Value.(*WrappedClient)
			client.borrowNum += 1
			p.mu.Unlock()
			//检测获取的client是否有异常，判断标准为client.clinet 不为空
			//检测client是否达到最大的调用次数，如果超过，就扔掉
			if client.client == nil || client.borrowNum >= 20 {
				p.mu.Lock()
				p.release()
				continue
			}
			return client, nil
		}
		// 获取不到并且没有超过最大链接数限制新建立链接
		if p.MaxActive == 0 || p.active < p.MaxActive {
			p.active += 1
			p.mu.Unlock()
			client, err := p.newThriftClient()
			if err != nil {
				p.mu.Lock()
				p.release()
				p.mu.Unlock()
				client = nil
			}
			// 为保证都有返回，即使有err也返回一个WrappedClient
			return &WrappedClient{
				client: client,
				p:      p,
				err:    err,
			}, err
		}

		if !p.Wait {
			p.mu.Unlock()
			err := errors.New("connection pool exhausted")
			return &WrappedClient{
				client: nil,
				p:      p,
				err:    err,
			}, err
		}

		if p.cond == nil {
			p.cond = sync.NewCond(&p.mu)
		}
		p.cond.Wait()
	}
}

// 将client放回去, 上层调用会有可能出err，此时把这个client 关掉
func (p *Pool) put(wrappedClient *WrappedClient) {
	p.mu.Lock()
	// 如果有错误直接丢弃掉
	if wrappedClient.err != nil {
		p.release()
		p.mu.Unlock()
		closeTransport(wrappedClient.client)
		return
	}
	// 放入到pool中
	p.idle.PushFront(wrappedClient)
	if p.idle.Len() <= p.MaxIdle {
		if p.cond != nil {
			p.cond.Signal()
		}
		p.mu.Unlock()
		return
	}
	// 如果超长就扔掉最后一个
	wrappedClient = p.idle.Remove(p.idle.Back()).(*WrappedClient)
	p.release()
	p.mu.Unlock()
	closeTransport(wrappedClient.client)
	return
}

func (p *Pool) CallWithRetry(name string, args ...interface{}) (res interface{}, err error) {
	client, _ := p.Get()
	defer client.Close()
	res, err = client.CallWithRetry(name, args...)
	return
}

func RegisterPool(poolname string, clientFactory interface{}) {
	poolConfig, _ := config.GlobalConfiger.Map(fmt.Sprintf("thriftpool::%s", poolname))
	field := fmt.Sprintf("thriftpool::%s", poolname)

	hosts := config.GlobalConfiger.Strings(
		fmt.Sprintf("%s::hosts", field))

	framed, _ := config.GlobalConfiger.Bool(
		fmt.Sprintf("%s::framed", field))

	maxIdle, _ := config.GlobalConfiger.Int(
		fmt.Sprintf("%s::max_idle", field))

	maxRetry, _ := config.GlobalConfiger.Int(
		fmt.Sprintf("%s::max_retry", field))

	withCommonHeader := config.GlobalConfiger.Bool(
		fmt.Sprintf("%s::with_common_header", field))

	GlobalThriftPool[poolname] = NewThriftPool(
		clientFactory,
		hosts, framed, maxIdle, uint(maxRetry), withComonHeader)

}

func init() {
	// 种子只初始化一次，用以保证生成的是随机化序列
	rand.Seed(time.Now().Unix())
}
