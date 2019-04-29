// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	AD_WORDS = []string{"t.me/", "t.cn/", "joinchat"}
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
	f := j.app.Features.AdDetection
	msg := j.req.Message

	if !f.Enabled ||
		!j.HasMessageContent() {
		return nil, nil
	}

	// detection of Telegram groups
	if isAd(&msg) {
		admins := []string{BDFL}

		if f.AdminAlert {
			admins = append(j.app.ChatAdmins[msg.Chat.Id])
		}

		for _, a := range admins {
			if msg.From.Username == a {
				return nil, nil
			}
		}

		j.app.Logger.WithFields(logrus.Fields{
			"chat":   msg.Chat.Id,
			"userId": msg.From.Id,
		}).Warn("Ad detected")

		adminsToNotify := strings.Join(admins, ", ")

		text := fmt.Sprintf(f.I18n[j.app.Lang].WarnMessage, adminsToNotify)

		req := &BotSendMsg{
			ChatId:           msg.Chat.Id,
			Text:             text,
			ParseMode:        ParseModeMarkdown,
			ReplyToMessageId: msg.MessageId,
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
				case <-time.After(time.Duration(TIME_TO_DELETE_REPLY_MSG+10) * time.Second):
					go j.DeleteMessage(&msg)
					go j.DeleteMessage(replyMsgBody)
				}
			}()
		}
	}

	return nil, nil
}
