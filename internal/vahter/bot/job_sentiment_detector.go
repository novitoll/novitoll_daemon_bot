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
	t := j.req.Message.Text

	if !f.Enabled || !j.HasMessageContent() {
		return nil, nil
	}

	// check if the entire message is emoji
	if t != "" && emojiRgxp.ReplaceAllString(t, "") == "" {
		emojiMsg = true
	}

	if emojiMsg || j.req.Message.Sticker.FileId != "" {
		j.app.Logger.WithFields(logrus.Fields{"userId": j.req.Message.From.Id}).
			Warn("Sentiment detected")
	} else {
		return nil, nil
	}

	// sentiment detected - warn user
	text := f.I18n[j.app.Lang].WarnMessage

	reply, err := j.SendMessage(text, j.req.Message.MessageId)
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
					go j.DeleteMessage(&j.req.Message)
				}
			}
		}()
	}

	return reply, err
}
