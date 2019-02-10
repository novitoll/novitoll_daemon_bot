// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"github.com/sirupsen/logrus"
)

type Job struct {
	req *BotInReq
	app         *App
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
		ChatId:                j.req.Message.Chat.Id,
		Text:                  replyText,
		ParseMode:             ParseModeMarkdown,
		ReplyToMessageId:      replyMsgId,
		ReplyMarkup:           &BotForceReply{ForceReply: false,
								Selective: true},
	}
	return req.SendMsg(j.app)
}