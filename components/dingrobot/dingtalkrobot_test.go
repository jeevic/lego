package dingrobot

import (
	"testing"
)

func TestDingTalkRobot_TextMsg_Send(t *testing.T) {
	url := "https://oapi.dingtalk.com/robot/send?access_token"
	secret := ""
	robot, _ := NewDingDingTalkRobot(url, secret)

	/*msg := NewTextMsg()

	msg.Text.Content = "这是一个测试aaaa"
	msg.At.AtMobiles = []string{"13683506199"}
	msg.At.AtUserIds = []string{"lx6g9h8"}*/

	msg := NewLinkMsg()

	msg.Link.MessageUrl = "https://www.baidu.com/"
	msg.Link.PicUrl = "https://www.baidu.com/img/PCfb_5bf082d29588c07f842ccde3f97243ea.png"
	msg.Link.Text = "这是一个寂寞的夜， 下着有点伤心的雨"
	msg.Link.Title = "让累化作伤心雨"

	res, err := robot.Send(msg)
	if err != nil {
		t.Log(err)
	}
	t.Log(res)

}

func TestDingTalkRobot_BulidSign(t *testing.T) {
	url := "https://oapi.dingtalk.com/robot/send?access_token="
	secret := ""
	robot, _ := NewDingDingTalkRobot(url, secret)

	var millSeconds int64 = 1618228844084

	s := robot.buildSign(millSeconds)
	t.Log(s)
}
