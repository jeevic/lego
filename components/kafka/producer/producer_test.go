package producer

import (
	"fmt"
	"testing"
)

func TestNewKafkaProducer(t *testing.T) {
	producerSetting := NewSetting()
	producerSetting.Hosts = []string{"10.103.17.53:9092"}
	producerSetting.Topic = "contech_image_uploaded1"
	producerSetting.ReturnSuccess = true
	producerSetting.RequiredAcks = 0
	producer, _ := NewKafkaProducer(producerSetting)
	partition, offset, err := producer.SendMsgSync("contech_image_uploaded1", "111", "22222")
	if err != nil {
		fmt.Println(err.Error())
	}
	result := fmt.Sprintf("send msg success  partition：%d  offset：%d\n", partition, offset)
	fmt.Println(result)
	//
	//partition, offset, err = producer.SendMsgSync("contech_image_uploaded1", "111", "3333333")
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//result = fmt.Sprintf("send msg success  partition：%d  offset：%d\n", partition, offset)
	//fmt.Println(result)
}
