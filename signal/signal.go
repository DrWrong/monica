package signal

type Subscriber func(publisher string, args ...interface{})


//同步消息 observer pattern
type Signal struct {
	Subscribers []Subscriber
}

func (signal *Signal) RegisterSubscribers(subscriber Subscriber) {
	signal.Subscribers = append(signal.Subscribers, subscriber)
}

func (signal *Signal) Emit(publisher string, args ...interface{}){
	for _, subscriber := range signal.Subscribers {
		subscriber(publisher, args...)
	}
}
