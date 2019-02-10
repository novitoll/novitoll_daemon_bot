// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"time"

	"github.com/sirupsen/logrus"
)

func JobStickersDetector(j *Job) (interface{}, error) {
	if !j.app.Features.StickersDetection.Enabled ||
		!j.HasMessageContent() || j.req.Message.Sticker.FileId == "" {
		return nil, nil
	}

	j.app.Logger.WithFields(logrus.Fields{
		"userId": j.req.Message.From.Id,
	}).Warn("Sticker detected")

	text := j.app.Features.StickersDetection.I18n[j.app.Lang].WarnMessage

	reply, err := j.SendMessage(text, j.req.Message.MessageId)
	if err != nil {
		return false, err
	}

	if reply != nil {
		// cleanup reply messages
		go func() {
			select {
			case <-time.After(time.Duration(TIME_TO_DELETE_REPLY_MSG+10) * time.Second):
				go j.DeleteMessage(&j.req.Message)
				go j.DeleteMessage(reply)
			}
		}()
	}

	return reply, err
}
