package dingrobot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jeevic/lego/components/httplib"
)

const MsgTypeText = "text"
const MsgTypeMarkdown = "markdown"
const MsgTypeLink = "link"

type DingTalkRobot struct {
	//钉钉发送地址 格式如:https://oapi.dingtalk.com/robot/send?access_token=1122344
	Url string
	//地址密钥 安全设置中加签 加签
	Secret string
}

func NewDingDingTalkRobot(url string, secret string) (*DingTalkRobot, error) {
	if len(url) < 1 {
		return nil, errors.New("url must not empty")
	}
	if len(secret) < 1 {
		return nil, errors.New("secret must not empty")
	}
	return &DingTalkRobot{
		Url:    url,
		Secret: secret,
	}, nil
}

//钉钉签名
func (d *DingTalkRobot) buildSign(millSecond int64) string {
	str := fmt.Sprintf("%d\n%s", millSecond, d.Secret)
	strSecret := hmacSha256(str, d.Secret)
	return url.QueryEscape(base64.StdEncoding.EncodeToString([]byte(strSecret)))
}

//发送ding talk
func (d *DingTalkRobot) Send(msg interface{}) (ResponseResult, error) {
	var res ResponseResult
	milSecond := time.Now().UnixNano() / 1e6
	sign := d.buildSign(milSecond)

	byts, err := json.Marshal(msg)
	if err != nil {
		return res, err
	}

	urlPath := fmt.Sprintf("%s&timestamp=%d&sign=%s", d.Url, milSecond, sign)
	h := httplib.Post(urlPath).Header("Content-Type", "application/json").SetTimeout(5*time.Second, 5*time.Second).Body(byts)
	err = h.ToJSON(&res)
	if err != nil {
		return res, err
	}
	return res, nil
}

//TextMsg 类型文档类型
type TextMsg struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At struct {
		AtMobiles []string `json:"atMobiles"`
		AtUserIds []string `json:"atUserIds"`
		IsAtAll   bool     `json:"isAtAll"`
	} `json:"at"`
}

func NewTextMsg() TextMsg {
	return TextMsg{
		MsgType: MsgTypeText,
	}
}

//link类型
type LinkMsg struct {
	MsgType string `json:"msgtype"`
	Link    struct {
		Text       string `json:"text"`
		Title      string `json:"title"`
		PicUrl     string `json:"picUrl"`
		MessageUrl string `json:"messageUrl"`
	} `json:"link"`
}

func NewLinkMsg() LinkMsg {
	return LinkMsg{
		MsgType: MsgTypeLink,
	}
}

//markdown类型
type MarkdownMsg struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown"`
	At struct {
		AtMobiles []string `json:"atMobiles"`
		AtUserIds []string `json:"atUserIds"`
		IsAtAll   bool     `json:"isAtAll"`
	} `json:"at"`
}

func NewMarkdownMsg() MarkdownMsg {
	return MarkdownMsg{
		MsgType: MsgTypeMarkdown,
	}
}

//返回结果
type ResponseResult struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func hmacSha256(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	h.Sum(nil)
	return string(h.Sum(nil))
}
