package consumer

import (
	"context"
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	"go.uber.org/atomic"

	"github.com/jeevic/lego/pkg/app"
)

var wg sync.WaitGroup

type Consumer struct {
	config             *sarama.Config
	setting            *setting
	groups             []sarama.ConsumerGroup
	partitionConsumers []sarama.PartitionConsumer
	stopFlag           *atomic.Bool //0标识 未关闭
}
type setting struct {
	Hosts       []string
	Topic       string
	GroupId     string
	Offset      int64
	AutoCommit  bool
	MaxRetry    int
	ReturnError bool
}

func NewSetting() *setting {
	s := &setting{}
	s.Offset = sarama.OffsetNewest
	s.MaxRetry = 3
	s.AutoCommit = true
	s.ReturnError = false
	return s
}

func NewConsumer(setting *setting) (*Consumer, error) {
	config := buildConsumerConfig(setting)
	return &Consumer{config, setting, make([]sarama.ConsumerGroup, 0), make([]sarama.PartitionConsumer, 0), atomic.NewBool(false)}, nil
}

func buildConsumerConfig(setting *setting) *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = setting.Offset
	config.Consumer.Offsets.Retry.Max = setting.MaxRetry
	config.Consumer.Offsets.AutoCommit.Enable = setting.AutoCommit
	config.Consumer.Return.Errors = setting.ReturnError
	return config
}

func (consumer *Consumer) ConsumerMsg(topic string, f func(msg string)) error {
	consumerClient, err := sarama.NewConsumer(consumer.setting.Hosts, consumer.config)
	if err != nil {
		return err
	}

	if len(topic) == 0 && len(consumer.setting.Topic) > 0 {
		topic = consumer.setting.Topic
	}
	partitions, err := consumerClient.Partitions(topic)
	if err != nil {
		return err
	}
	for _, p := range partitions {
		partitionConsumer, err := consumerClient.ConsumePartition(topic, p, consumer.setting.Offset)
		if err != nil {
			continue
		}
		consumer.partitionConsumers = append(consumer.partitionConsumers, partitionConsumer)
		wg.Add(1)
		for msg := range partitionConsumer.Messages() {
			f(string(msg.Value))
		}
		wg.Done()
	}
	wg.Wait()
	return nil
}

func (consumer *Consumer) ConsumerGroupMsg(topic string, groupId string, handler sarama.ConsumerGroupHandler) error {
	if len(topic) == 0 && len(consumer.setting.Topic) > 0 {
		topic = consumer.setting.Topic
	}
	if len(groupId) == 0 && len(consumer.setting.GroupId) > 0 {
		groupId = consumer.setting.GroupId
	}
	group, err := sarama.NewConsumerGroup(consumer.setting.Hosts, groupId, consumer.config)
	if err != nil {
		return err
	}
	consumer.groups = append(consumer.groups, group)
	defer func() { _ = group.Close() }()

	// Track errors
	go func() {
		for err := range group.Errors() {
			app.App.GetLogger().Errorf("ERROR %s", err.Error())
		}
	}()
	ctx := context.Background()
	for !consumer.stopFlag.Load() {
		topics := strings.Split(topic, ",")
		err := group.Consume(ctx, topics, handler)
		if err != nil && !consumer.stopFlag.Load() {
			app.App.GetLogger().Errorf("ERROR %s", err.Error())
		}
	}
	return nil
}
func (consumer *Consumer) Close() {
	consumer.stopFlag.Store(true)
	for _, item := range consumer.groups {
		if item != nil {
			item.Close()
		}
	}
	for _, item := range consumer.partitionConsumers {
		if item != nil {
			item.Close()
		}
	}
}
