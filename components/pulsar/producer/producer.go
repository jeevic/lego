package producer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

type Producer struct {
	producer pulsar.Producer
	client   pulsar.Client
	setting  *setting
}

type setting struct {
	Hosts             string
	Topic             string
	OperationTimeout  time.Duration
	ConnectionTimeout time.Duration
	Token             string
}

func NewSetting() *setting {
	s := &setting{}
	s.OperationTimeout = 30 * time.Second
	s.ConnectionTimeout = 30 * time.Second
	return s
}

func NewProducer(setting *setting) (*Producer, error) {
	options := pulsar.ClientOptions{}
	options.URL = setting.Hosts
	options.OperationTimeout = setting.OperationTimeout
	options.ConnectionTimeout = setting.ConnectionTimeout
	if len(setting.Token) > 0 {
		options.Authentication = pulsar.NewAuthenticationToken(setting.Token)
	}
	client, err := pulsar.NewClient(options)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("create pulsar client error:%s", err.Error()))
	}
	producerOptions := pulsar.ProducerOptions{Topic: setting.Topic}
	producer, err := client.CreateProducer(producerOptions)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("create pulsar producer error:%s", err.Error()))
	}
	return &Producer{producer: producer, setting: setting}, nil
}
func (pulsarProducer *Producer) SendMsgSync(key string, value string) (msgId pulsar.MessageID, err error) {
	return pulsarProducer.SendMsgDelay(key, value, -1)
}

func (pulsarProducer *Producer) SendMsgDelay(key string, value string, delayAfter time.Duration) (msgId pulsar.MessageID, err error) {
	msg := pulsar.ProducerMessage{Payload: []byte(value), Key: key}
	if delayAfter > 0 {
		msg.DeliverAfter = delayAfter
	}
	msgId, err = pulsarProducer.producer.Send(context.Background(), &msg)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("send message error:%s", err.Error()))
	}
	return msgId, nil
}

func (pulsarProducer *Producer) Close() {
	pulsarProducer.producer.Close()
	pulsarProducer.client.Close()
}
