package main

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xpzouying/keaimao"
)

func doSendToFeishu(text string) {
	feishuAPI := viper.GetString("feishu.api")

	if feishuAPI == "" {
		return
	}

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

func doSendToWeixinRobot(text string) {
	var (
		keaimaoAPI = viper.GetString("weixin.api")
		robotid    = viper.GetString("weixin.robot_wxid")
		to         = viper.GetString("weixin.to")
	)

	if keaimaoAPI == "" {
		return
	}

	mao := keaimao.NewRobot(robotid, keaimao.WithSendAPI(keaimaoAPI))
	if err := mao.SendGroupMessage(context.Background(), to, robotid, text); err != nil {
		logrus.Errorf("send weixin group error: %v", err)
		return
	}
}
