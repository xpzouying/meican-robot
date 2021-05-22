package main

import (
	"bytes"
	"context"
	"encoding/json"
	"html"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xpzouying/gomeican"
)

var (
	httpclient = &http.Client{Timeout: 5 * time.Second}
)

type M map[string]interface{}

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	meicanToken := viper.GetString("robot.token")
	if len(meicanToken) == 0 {
		panic("empty token, set token first")
	}

	meican := gomeican.NewMeican(meicanToken)

	d := time.Now()
	dinnerOrders, err := meican.GetOrderList(context.Background(), d)
	if err != nil {
		logrus.Errorf("get order list error: %v", err)
		os.Exit(-1)
	}

	msgs := makeSendMsg(dinnerOrders)

	sendMessages(msgs)
}

func sendMessages(msgs []string) {

	if count := len(msgs); count == 0 {
		logrus.Infof("no msgs to send")
		return
	} else if count != 2 {
		logrus.Warnf("invalid msgs length: len(msgs)=%d", count)
		return
	}

	if could := couldSendMsg(); !could {
		logrus.Warnf("couldn't send msg by policy")
		return
	}

	doSend(msgs)
}

func doSend(msgs []string) {
	h := time.Now().Local().Hour()

	for _, msg := range msgs {
		doSendToFeishu(msg)
	}

	if h >= 7 && h <= 9 { // 上午可以发送的时间：早上7点 到 早上9点
		logrus.Infof("send: %s", msgs[0])
	} else if h >= 15 && h <= 17 { // 下午的发送时间: 3 - 5
		logrus.Infof("send: %s", msgs[1])
	} else {
		logrus.Warnf("invalid time to send: %s", msgs)
	}
}

func doSendToFeishu(text string) {
	feishuAPI := viper.GetString("feishu.api")

	msg := M{
		"msg_type": "text",
		"content": M{
			"text": text,
		},
	}

	data, _ := json.Marshal(msg)
	var resp struct {
		StatusCode    int
		StatusMessage string
	}

	if err := postWithDecode(context.Background(), feishuAPI, bytes.NewReader(data), &resp); err != nil {
		logrus.Errorf("post feishu error: %v", err)
		return
	}

	if resp.StatusCode != 0 {
		logrus.Errorf("post feishu error: status_code=%v status_msg:%s", resp.StatusCode, resp.StatusMessage)
		return
	}
}

// 发送的一些策略
func couldSendMsg() bool {
	debug := viper.GetBool("robot.debug")

	if debug {
		return true
	}

	// 有效的发送时间段
	now := time.Now()
	switch wd := now.Local().Weekday(); wd {
	case time.Saturday, time.Sunday:
		return false
	}

	return true
}

func makeSendMsg(orders []gomeican.DinnerOrder) []string {

	ss := make([]string, 0, len(orders))

	for _, order := range orders {
		if msg := makeOneSendMsg(order); msg != "" {
			ss = append(ss, msg)
		}
	}

	return ss
}

func makeOneSendMsg(order gomeican.DinnerOrder) string {
	buf := new(bytes.Buffer)
	t := template.Must(template.New("msg").Parse(msgTemplate))

	if err := t.Execute(buf, order); err != nil {
		return ""
	}

	s := buf.String()
	return html.UnescapeString(s)
}

const msgTemplate = `
{{.TimeInfo.Title}}预定时间（{{.TimeInfo.OpeningTime.OpenTime}} - {{.TimeInfo.OpeningTime.CloseTime}}）
{{range $info := .RestaurantFoodInfos}}
{{$info.RestaurantInfo.Name}} - 数量：{{$info.RestaurantInfo.AvailableDishCount}} / {{$info.RestaurantInfo.DishLimit}}
  {{range $foodInfo := $info.FoodList}}
  🐈 {{$foodInfo.Name}}
  {{end}}
{{end}}
`
