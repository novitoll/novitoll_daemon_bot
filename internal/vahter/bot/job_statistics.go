// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/novitoll/novitoll_daemon_bot/internal/utils"
	"github.com/sirupsen/logrus"
)

const (
	FLOOD_TIME_INTERVAL     = 10
	FLOOD_MAX_ALLOWED_MSGS  = 3
	FLOOD_MAX_ALLOWED_WORDS = 500
	CONT_MSGS_ALLOWED       = 20
	CONT_USER_MSG_ALLOWED   = 3
)

// formula 1. (Incremental average) M_n = M_(n-1) + ((A_n - M_(n-1)) / n),
// where M_n = total mean, n = count of records, A = the array of elements

type UserMessageStats struct {
	FloodMsgsLength   []int
	AllMsgsCount      int
	LastMsgTime       int64
	SinceLastMsg      int
	MeanAllMsgsLength int
	Dates             []int64
	Username          string
}

func JobMsgStats(j *Job) (interface{}, error) {
	if !j.app.Features.MsgStats.Enabled || !j.HasMessageContent() {
		return nil, nil
	}

	// 1. get the s. s
	s := UserStatistics[j.req.Message.From.Id]
	W := len(strings.Fields(j.req.Message.Text))
	t0 := time.Now().Unix()

	if s == nil {
		// 2.1 init the user s
		s = &UserMessageStats{
			FloodMsgsLength:   []int{W},
			AllMsgsCount:      0,
			LastMsgTime:       t0,
			SinceLastMsg:      0,
			MeanAllMsgsLength: 0,
			Username:          j.req.Message.From.Username,
		}
	} else {
		// 2.2 update the user s
		s.Dates = append(s.Dates, j.req.Message.Date)
		s.FloodMsgsLength = append(s.FloodMsgsLength, W)
		s.AllMsgsCount += 1

		// Ref:formula 1.
		s.MeanAllMsgsLength += ((W - s.MeanAllMsgsLength) / s.AllMsgsCount)

		s.SinceLastMsg = int(time.Since(time.Unix(s.LastMsgTime, 0)).Seconds())
		s.LastMsgTime = t0
	}

	// 3. Detect if user has been flooding for last TIME_INTERVAL seconds
	// add here the condition with the MeanMsgLength within TIME_INTERVAL
	err := floodDetection(j, s)

	return s, err
}

func floodDetection(j *Job, s *UserMessageStats) error {
	var isFlood bool
	var replyText []string

	// for short reference
	template := j.app.Features.MsgStats.I18n[j.app.Lang]

	if s.SinceLastMsg <= FLOOD_TIME_INTERVAL &&
		len(s.FloodMsgsLength) >= FLOOD_MAX_ALLOWED_MSGS {

		j.app.Logger.WithFields(logrus.Fields{
			"userId": j.req.Message.From.Id,
		}).Warn("User is flooding")

		s.FloodMsgsLength = []int{}
		isFlood = true
		replyText = append(replyText, fmt.Sprintf(template.WarnMessageTooFreq,
			FLOOD_TIME_INTERVAL, FLOOD_MAX_ALLOWED_MSGS))
	}

	if allWordsCount := utils.SumSliceInt(s.FloodMsgsLength); allWordsCount >= FLOOD_MAX_ALLOWED_WORDS {

		isFlood = true
		replyText = append(replyText, fmt.Sprintf(template.WarnMessageTooLong,
			FLOOD_MAX_ALLOWED_WORDS))

		// notify admins
		var admins []string

		for _, a := range j.app.ChatAdmins[j.req.Message.Chat.Id] {
			admins = append(admins, fmt.Sprintf("@%s", a))
		}
		replyText = append(replyText, fmt.Sprintf(". CC: %s",
			strings.Join(admins, ", ")))
	}

	if isFlood {
		// notify user about the flood limit
		txt := strings.Join(replyText, ". ")

		reply, err := j.SendMessage(txt, j.req.Message.MessageId)
		if err != nil {
			return err
		}

		if reply != nil {
			// cleanup reply messages
			go func() {
				select {
				case <-time.After(time.Duration(
					TIME_TO_DELETE_REPLY_MSG+10) * time.Second):
					j.DeleteMessage(reply)
				}
			}()
		}
	}

	if s.SinceLastMsg > FLOOD_TIME_INTERVAL {
		s.FloodMsgsLength = []int{}
	}

	// 4. update the user s map
	UserStatistics[j.req.Message.From.Id] = s
	return nil
}
