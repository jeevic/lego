package consumer

import (
	"fmt"
	"testing"

	"github.com/Shopify/sarama"
)

type exampleConsumerGroupHandler struct{}

func (exampleConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (exampleConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h exampleConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("Message topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)
		fmt.Println(string(msg.Value))
		sess.MarkMessage(msg, "")
	}
	return nil
}

func TestNewConsumer(t *testing.T) {
	consumerSetting := NewSetting()
	consumerSetting.Hosts = []string{"10.136.40.11:9092", "10.136.40.12:9092", "10.136.40.13:9092"}
	consumerSetting.Topic = "contech_image_uploaded_test"
	consumerSetting.ReturnError = true
	kafkaConsumer, err := NewConsumer(consumerSetting)
	if err != nil {
		fmt.Println(err.Error())
	}
	handler := exampleConsumerGroupHandler{}
	kafkaConsumer.ConsumerGroupMsg(consumerSetting.Topic, "image-pipeline", handler)
}

func TestNewKafkaConsumer(t *testing.T) {

}

func consumerMsgFunc(msg string) {
	fmt.Println("consumer msg :" + msg)
}
