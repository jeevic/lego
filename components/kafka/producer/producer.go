package producer

import (
	"errors"
	"fmt"
	"time"

	"github.com/Shopify/sarama"

	"github.com/jeevic/lego/pkg/app"
)

type Producer struct {
	client  sarama.Client
	setting *setting
}

type setting struct {
	Hosts         []string
	Topic         string
	ReturnSuccess bool
	ReturnError   bool
	RequiredAcks  int
	Timeout       int
	MaxRetry      int
}

func NewSetting() *setting {
	s := &setting{}
	s.MaxRetry = 3
	s.RequiredAcks = 1
	s.ReturnSuccess = true
	s.ReturnError = true
	s.Timeout = 10
	return s
}

func NewKafkaProducer(producerSetting *setting) (*Producer, error) {
	config := buildProducerConfig(producerSetting)
	client, err := sarama.NewClient(producerSetting.Hosts, config)
	if err != nil {
		return nil, err
	}
	return &Producer{client, producerSetting}, nil
}

func buildProducerConfig(producerSetting *setting) *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.Retry.Max = producerSetting.MaxRetry
	switch producerSetting.RequiredAcks {
	case -1:
		config.Producer.RequiredAcks = sarama.WaitForAll
	case 0:
		config.Producer.RequiredAcks = sarama.NoResponse
	case 1:
		config.Producer.RequiredAcks = sarama.WaitForLocal
	}
	config.Producer.Return.Successes = producerSetting.ReturnSuccess
	config.Producer.Return.Errors = producerSetting.ReturnError
	if producerSetting.Timeout > 0 {
		config.Producer.Timeout = time.Duration(producerSetting.Timeout) * time.Second
	}
	return config
}

func (kafkaProducer *Producer) SendMsgSync(topic string, key string, value string) (partition int32, offset int64, err error) {
	if len(topic) == 0 && len(kafkaProducer.setting.Topic) > 0 {
		topic = kafkaProducer.setting.Topic
	}
	syncProducer, err := sarama.NewSyncProducerFromClient(kafkaProducer.client)
	if err != nil {
		return int32(-1), int64(-1), errors.New(fmt.Sprintf(" create sync producer error:%s", err.Error()))
	}
	defer syncProducer.Close()
	msg := sarama.ProducerMessage{Topic: topic, Key: sarama.StringEncoder(key), Value: sarama.ByteEncoder(value)}
	partition, offset, err = syncProducer.SendMessage(&msg)
	if err != nil {
		return int32(-1), int64(-1), errors.New(fmt.Sprintf("send message error:%s", err.Error()))
	}
	return partition, offset, nil
}

func (kafkaProducer *Producer) SendMsgASync(topic string, key string, value string) error {
	if len(topic) == 0 && len(kafkaProducer.setting.Topic) > 0 {
		topic = kafkaProducer.setting.Topic
	}
	asyncProducer, err := sarama.NewAsyncProducerFromClient(kafkaProducer.client)
	if err != nil {
		return errors.New(fmt.Sprintf(" create aSync producer error:%s", err.Error()))
	}
	defer asyncProducer.Close()
	msg := sarama.ProducerMessage{Topic: topic, Key: sarama.StringEncoder(key), Value: sarama.ByteEncoder(value)}
	asyncProducer.Input() <- &msg
	select {
	case suc := <-asyncProducer.Successes():
		app.App.GetLogger().Debugf("offset: %d,  timestamp: %s", suc.Offset, suc.Timestamp.String())
		return nil
	case fail := <-asyncProducer.Errors():
		app.App.GetLogger().Errorf("err: %s\n", fail.Err.Error())
		return fail
	}
}
func (kafkaProducer *Producer) Close() {
	kafkaProducer.client.Close()
}
