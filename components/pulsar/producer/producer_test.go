package producer

import (
	"fmt"
	"testing"
)

func TestProducer_SendMsgSync(t *testing.T) {

	msg := "{\"doc_id\":\"0dCmd5QN\",\"ctype\":\"news\",\"Wm_id\":2356884,\"image_urls\":[\"YD_cnt_13_019hBQABWJOs\"],\"date\":\"2022-03-31 17:24:26\",\"retry_times\":1,\"retry_date\":\"2022-03-31 17:24:26\",\"request_id\":\"bc0bbaf8-22fb-43f4-b2ef-0821e0bca819\",\"image_type\":\"image_all\",\"version\":1648718666830,\"images_process_infos\":[{\"image_id\":\"YD_cnt_13_019hBQABWJOs\",\"image_url\":\"https://i1.go2yd.com/image.php?url=YD_cnt_13_019hBQABWJOs\",\"index_in_news\":0,\"code\":2,\"Error\":\"\",\"image_caption\":\"\"}],\"extend\":\"{\\\"mode\\\": \\\"prod\\\"}\"}"
	key := "0dCmd5QN"

	setting := NewSetting()
	setting.Hosts = "pulsar://10.103.17.55:6650,10.120.187.33:6650,10.120.187.34:6650"
	setting.Topic = "public/content/contech_markthal_warehouse_to_image_retry_test"
	setting.Token = ""
	producer, _ := NewProducer(setting)
	msgId, _ := producer.SendMsgSync(key, msg)
	fmt.Printf("send msg:%v", msgId)
}
