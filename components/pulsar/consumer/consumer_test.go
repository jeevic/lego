package consumer

import (
	"fmt"
	"testing"

	"github.com/apache/pulsar-client-go/pulsar"
)

var num = 1

func TestConsumer_ConsumerMsg(t *testing.T) {
	setting := NewSetting()
	setting.Hosts = "pulsar://10.103.17.55:6650,10.120.187.33:6650,10.120.187.34:6650"
	setting.Topic = "public/content/contech_markthal_warehouse_to_image_retry_test"
	setting.Subscription = "image_exchange"
	setting.Token = ""
	consumer, _ := NewConsumer(setting)
	consumer.ConsumerMsg(consumer.handlerMsg)

}

func (consumer *Consumer) handlerMsg(msg pulsar.Message) {
	fmt.Printf(
		"Received message  msgId: %v -- content: '%s'\n",
		msg.ID(),
		string(msg.Payload()),
	)
	if num == 1 {
		consumer.NackMsg(msg)
	} else {
		consumer.AckMsg(msg)
	}
	num = num + 1
}
