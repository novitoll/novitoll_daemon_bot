// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	AD_WORDS = []string{"t.me/", "joinchat"}
)

func isAd(msg *BotIngressRequestMessage) bool {
	var isMsgAd bool = false
	contexts := []string{msg.Text, msg.Caption}
	for _, s := range contexts {
		if isMsgAd {
			break
		}
		for _, ad := range AD_WORDS {
			if strings.Contains(s, ad) {
				isMsgAd = true
			}
		}
	}
	return isMsgAd
}

func JobAdDetector(job *Job) (interface{}, error) {
	if !job.app.Features.AdDetection.Enabled || !job.HasMessageContent() {
		return nil, nil
	}

	// detection of Telegram groups
	if isAd(&job.ingressBody.Message) {
		admins := job.app.ChatAdmins[job.ingressBody.Message.Chat.Id]

		for _, a := range admins {
			if fmt.Sprintf("@%s", job.ingressBody.Message.From.Username) == a {
				return nil, nil
			}
		}

		job.app.Logger.WithFields(logrus.Fields{
			"userId": job.ingressBody.Message.From.Id,
		}).Warn("Ad detected: Telegram group")

		adminsToNotify := strings.Join(admins, ", ")
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
		replyMsgBody, err := botEgressReq.EgressSendToTelegram(job.app)
		if err != nil {
			return false, err
		}

		if replyMsgBody != nil {
			// cleanup reply messages
			go func() {
				select {
				case <-time.After(time.Duration(TIME_TO_DELETE_REPLY_MSG+10) * time.Second):
					go job.DeleteMessage(&job.ingressBody.Message)
					go job.DeleteMessage(replyMsgBody)
				}
			}()
		}
	}

	return nil, nil
}
