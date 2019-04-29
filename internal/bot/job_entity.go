// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type Job struct {
	req *BotInReq
	app *App
}

func (j *Job) HasMessageContent() bool {
	return j.req.Message.Text != "" ||
		j.req.Message.Sticker.FileId != "" ||
		j.req.Message.Caption != ""
}

func (j *Job) DeleteMessage(resp *BotInReqMsg) (interface{}, error) {
	j.app.Logger.WithFields(logrus.Fields{"id": resp.MessageId}).Info("Deleting a reply message")

	req := &BotDeleteMsg{
		ChatId:    resp.Chat.Id,
		MessageId: resp.MessageId,
	}

	_, err := req.DeleteMsg(j.app)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (j *Job) SendMessage(replyText string, replyMsgId int) (*BotInReqMsg, error) {
	req := &BotSendMsg{
		ChatId:           j.req.Message.Chat.Id,
		Text:             replyText,
		ParseMode:        ParseModeMarkdown,
		ReplyToMessageId: replyMsgId,
		ReplyMarkup: &BotForceReply{ForceReply: false,
			Selective: true},
	}
	return req.SendMsg(j.app)
}

func (j *Job) SaveInRedis(redisConn *redis.Client, k string, v interface{}, t int) {
	err := redisConn.Set(k, v, time.Duration(t)*time.Second).Err()
	if err != nil {
		txt := fmt.Sprintf("Could not save %s in redis", k)
		j.app.Logger.Warn(txt)
	}
}

func (j *Job) GetBatchFromRedis(redisConn *redis.Client, k string, limit int) interface{} {
	var cursor uint64

	keys, _, err := redisConn.Scan(cursor, k, int64(limit)).Result()
	if err != nil {
		txt := fmt.Sprintf("Could not scan batch %s in redis", k)
		j.app.Logger.Warn(txt)
		return nil
	} else {
		return keys
	}
}
