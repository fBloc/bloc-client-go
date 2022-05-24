package event

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

func init() {
	var _ DomainEvent = &ClientRunFunction{}
}

type ClientRunFunction struct {
	FunctionRunRecordID string
	ClientName          string
	deliveryTag         uint64
}

func (event *ClientRunFunction) Topic() string {
	return "function_client_run_consumer." + event.ClientName
}

func (event *ClientRunFunction) DeliveryTag() uint64 {
	return event.deliveryTag
}

// Marshal .
func (event *ClientRunFunction) Marshal() ([]byte, error) {
	return json.Marshal(event)
}

// Unmarshal .
func (event *ClientRunFunction) Unmarshal(data *amqp.Delivery) (err error) {
	err = json.Unmarshal(data.Body, event)
	event.deliveryTag = data.DeliveryTag
	return
}

// Identity
func (event *ClientRunFunction) Identity() string {
	return event.FunctionRunRecordID
}
