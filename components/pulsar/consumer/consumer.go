package consumer

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

var wg sync.WaitGroup

type Consumer struct {
	client   pulsar.Client
	Consumer pulsar.Consumer
	setting  *setting
}
type setting struct {
	Hosts             string
	Topic             string
	Subscription      string //可以理解为kafka的groupId
	OperationTimeout  time.Duration
	ConnectionTimeout time.Duration
	Token             string
	Type              pulsar.SubscriptionType //默认为share模式
	ChanSize          int
}

func NewSetting() *setting {
	s := &setting{}
	s.OperationTimeout = 30 * time.Second
	s.ConnectionTimeout = 30 * time.Second
	s.Type = pulsar.Shared
	s.ChanSize = 10
	return s
}

func NewConsumer(setting *setting) (*Consumer, error) {
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
	channel := make(chan pulsar.ConsumerMessage, setting.ChanSize)
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            setting.Topic,
		SubscriptionName: setting.Subscription,
		MessageChannel:   channel,
		Type:             setting.Type})
	return &Consumer{client, consumer, setting}, nil
}

func (consumer *Consumer) ConsumerMsg(f func(msg pulsar.Message)) {
	for cm := range consumer.Consumer.Chan() {
		msg := cm.Message
		f(msg)
	}
}

// Acknowledge the message so that it can be deleted by the message broker
func (consumer *Consumer) AckMsg(msg pulsar.Message) {
	consumer.Consumer.Ack(msg)
}

// Message failed to process, redeliver later default 1 minute
func (consumer *Consumer) NackMsg(msg pulsar.Message) {
	consumer.Consumer.Nack(msg)
}

func (consumer *Consumer) Close() {
	consumer.Consumer.Close()
	consumer.client.Close()
}
