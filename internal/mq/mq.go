package mq

import "github.com/streadway/amqp"

type MsgQueue interface {
	Pub(topic string, data []byte) error
	Pull(topic, pullerTag string, respMsgByteChan chan *amqp.Delivery) error
	Ack(deliveryTag uint64) error
}
