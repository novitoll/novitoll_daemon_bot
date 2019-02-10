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

func isAd(msg *BotInReqMsg) bool {
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

func JobAdDetector(j *Job) (interface{}, error) {
	if !j.app.Features.AdDetection.Enabled || 
		!j.HasMessageContent() {
		return nil, nil
	}

	// detection of Telegram groups
	if isAd(&j.req.Message) {
		admins := j.app.ChatAdmins[j.req.Message.Chat.Id]

		for _, a := range admins {
			if fmt.Sprintf("@%s", j.req.Message.
				From.Username) == a {
				return nil, nil
			}
		}

		j.app.Logger.WithFields(logrus.Fields{
			"userId": j.req.Message.From.Id,
		}).Warn("Ad detected: Telegram group")

		adminsToNotify := strings.Join(admins, ", ")
		
		text := fmt.Sprintf(j.app.Features.AdDetection.
			I18n[j.app.Lang].WarnMessage, adminsToNotify)

		req := &BotSendMsg{
			ChatId:                j.req.Message.Chat.Id,
			Text:                  text,
			ParseMode:             ParseModeMarkdown,
			ReplyToMessageId:      j.req.Message.MessageId,
			ReplyMarkup: &BotForceReply{
				ForceReply: false,
				Selective:  true,
			},
		}
		replyMsgBody, err := req.SendMsg(j.app)
		if err != nil {
			return false, err
		}

		if replyMsgBody != nil {
			// cleanup reply messages
			go func() {
				select {
				case <-time.After(time.Duration(TIME_TO_DELETE_REPLY_MSG +
					10) * time.Second):

					go j.DeleteMessage(&j.req.Message)
					go j.DeleteMessage(replyMsgBody)
				}
			}()
		}
	}

	return nil, nil
}