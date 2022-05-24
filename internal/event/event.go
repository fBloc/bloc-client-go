package event

import (
	"github.com/fBloc/bloc-client-go/internal/mq"
	"github.com/streadway/amqp"

	"github.com/pkg/errors"
)

type DomainEvent interface {
	Topic() string
	Marshal() ([]byte, error)
	Unmarshal(data *amqp.Delivery) (err error)
	Identity() string
	DeliveryTag() uint64
}

var needInitialMqInsAsEventChannelError = errors.New("lack init event mq rely")

// EventChannel 存取event的的通道，
// 由于是分布式的，故肯定需要引入消息队列中间件
type eventDriver struct {
	mqIns mq.MsgQueue
}

var (
	driver = eventDriver{}
)

func InjectMq(eventChannel mq.MsgQueue) {
	driver.mqIns = eventChannel
}

/*
ListenEvent 监听某项事件

对比PubEvent，为什么多了listenerTag参数呢？
因为发布是发布一种类型的事件，其不需要也不应该知道有哪些地方需要订阅此事件
也就是说对于同一个事件的发布，可能有多个订阅者，所以需要传入订阅者的标识
*/
func ListenEvent(
	event DomainEvent, listenerTag string,
	respEventChan chan DomainEvent,
) error {
	if driver.mqIns == nil {
		panic(needInitialMqInsAsEventChannelError)
	}

	deliveryChan := make(chan *amqp.Delivery)
	err := driver.mqIns.Pull(event.Topic(), listenerTag, deliveryChan)
	if err != nil {
		return errors.Wrap(err, "pull event failed")
	}

	go func() {
		for del := range deliveryChan {
			err = event.Unmarshal(del)
			if err != nil {
				panic(err)
			}
			respEventChan <- event
		}
	}()

	return nil
}

func AckEvent(
	event DomainEvent,
) error {
	if driver.mqIns == nil {
		panic(needInitialMqInsAsEventChannelError)
	}

	return driver.mqIns.Ack(event.DeliveryTag())
}
