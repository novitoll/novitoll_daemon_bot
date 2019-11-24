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

func (j *Job) SendMessageWReply(replyText string, replyMsgId int, reply interface{}) (*BotInReqMsg, error) {
	req := &BotSendMsg{
		ChatId:           j.req.Message.Chat.Id,
		Text:             replyText,
		ParseMode:        ParseModeMarkdown,
		ReplyToMessageId: replyMsgId,
		ReplyMarkup:      reply,
	}
	return req.SendMsg(j.app)
}

func (j *Job) onDeleteMessage(resp *BotInReqMsg, delay uint8) (interface{}, error) {
	// dirty hack to do the same function on either channel (fan-in pattern)
	select {
	case <-forceDeletion:
		return j.DeleteMessage(resp)
	case <-time.After(time.Duration(delay) * time.Second):
		return j.DeleteMessage(resp)
	}
}

func (j *Job) SendMessageWCleanup(text string, delay uint8, reply interface{}) (interface{}, error) {
	// Send message to user and delete own (bot's) message as cleanup
	msg := j.req.Message
	botEgressReq := &BotSendMsg{
		ChatId:           msg.Chat.Id,
		Text:             text,
		ParseMode:        ParseModeMarkdown,
		ReplyToMessageId: msg.MessageId,
		ReplyMarkup:      reply,
	}
	replyMsgBody, err := botEgressReq.SendMsg(j.app)
	if err != nil {
		return false, err
	}

	if replyMsgBody != nil {
		// cleanup reply messages
		go j.onDeleteMessage(replyMsgBody, delay)
	}

	return false, err
}

func (j *Job) KickChatMember(userId int, username string) (interface{}, error) {
	msg := j.req.Message
	t := time.Now().Add(time.Duration(j.app.Features.
		NewcomerQuestionnare.KickBanTimeout) * time.Second).Unix()

	j.app.Logger.WithFields(logrus.Fields{
		"chat":     msg.Chat.Id,
		"id":       userId,
		"username": username,
		"until":    t,
	}).Warn("Kicking a newcomer")

	botEgressReq := &BotKickChatMember{
		ChatId:    msg.Chat.Id,
		UserId:    userId,
		UntilDate: t,
	}
	return botEgressReq.KickChatMember(j.app)
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
