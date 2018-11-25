// SPDX-License-Identifier: GPL-2.0

package bot

import (
	"time"

	"github.com/sirupsen/logrus"
)

func JobStickersDetector(job *Job) (interface{}, error) {
	if !job.app.Features.StickersDetection.Enabled || !job.HasMessageContent() || job.ingressBody.Message.Sticker.FileId == "" {
		return nil, nil
	}

	job.app.Logger.WithFields(logrus.Fields{
		"userId": job.ingressBody.Message.From.Id,
	}).Warn("Sticker detected")

	text := job.app.Features.StickersDetection.I18n[job.app.Lang].WarnMessage

	botEgressReq := &BotEgressSendMessage{
		ChatId:                job.ingressBody.Message.Chat.Id,
		Text:                  text,
		ParseMode:             ParseModeMarkdown,
		DisableWebPagePreview: true,
		DisableNotification:   true,
		ReplyToMessageId:      job.ingressBody.Message.MessageId,
		ReplyMarkup: &BotForceReply{
			ForceReply: false,
			Selective:  true,
		},
	}

	replyMsgBody, err := botEgressReq.EgressSendToTelegram(job.app)
	if err != nil {
		return false, err
	}

	if replyMsgBody != nil {
		// cleanup reply messages
		go func() {
			job.DeleteMessage(&job.ingressBody.Message)

			select {
			case <-time.After(time.Duration(TIME_TO_DELETE_REPLY_MSG+10) * time.Second):
				job.DeleteMessage(replyMsgBody)
			}
		}()
	}

	return replyMsgBody, err
}
