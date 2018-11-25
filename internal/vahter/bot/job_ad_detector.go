// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func JobAdDetector(job *Job) (interface{}, error) {
	if !job.app.Features.AdDetection.Enabled || !job.HasMessageContent() {
		return nil, nil
	}

	// detection of Telegram groups
	if strings.Contains(job.ingressBody.Message.Text, "t.me/") {
		job.app.Logger.WithFields(logrus.Fields{
			"userId": job.ingressBody.Message.From.Id,
		}).Warn("Ad detected: Telegram group")

		adminsToNotify := strings.Join(job.app.Features.Administration.Admins, ", ")
		text := fmt.Sprintf(job.app.Features.AdDetection.I18n[job.app.Lang].WarnMessage, adminsToNotify)

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
		return botEgressReq.EgressSendToTelegram(job.app)
	}

	return nil, nil
}
