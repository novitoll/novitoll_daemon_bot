// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	SENTIMENTS = []string{")", ""}
	emojiRgxp  = regexp.MustCompile(`[\x{1F600}-\x{1F6FF}|[\x{2600}-\x{26FF}]|\)`)
)

func JobSentimentDetector(j *Job) (interface{}, error) {
	var emojiMsg bool

	f := j.app.Features.SentimentDetection
	msg := j.req.Message

	if !f.Enabled || !j.HasMessageContent() {
		return nil, nil
	}

	// check if the entire message is emoji
	if msg.Text != "" && emojiRgxp.ReplaceAllString(msg.Text, "") == "" {
		emojiMsg = true
	}

	if emojiMsg || msg.Sticker.FileId != "" {
		j.app.Logger.WithFields(logrus.Fields{"userId": msg.From.Id}).Warn("Sentiment detected")
	} else {
		return nil, nil
	}

	// sentiment detected - warn user
	replyText := f.I18n[j.app.Lang].WarnMessage

	reply, err := j.SendMessage(replyText, msg.MessageId)
	if err != nil {
		return false, err
	}

	if reply != nil {
		// cleanup reply messages
		go func() {
			select {
			case <-time.After(time.Duration(TIME_TO_DELETE_REPLY_MSG+10) * time.Second):
				go j.DeleteMessage(reply)

				if f.DeleteMessage {
					go j.DeleteMessage(&msg)
				}
			}
		}()
	}

	return reply, err
}
