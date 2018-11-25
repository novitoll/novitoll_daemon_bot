// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	FLOOD_TIME_INTERVAL    = 10
	FLOOD_MAX_ALLOWED_MSGS = 5
)

var (
	UserStatistics = make(map[int]*UserMessageStats)
)

// formula 1. (Incremental average) M_n = M_(n-1) + ((A_n - M_(n-1)) / n), where M_n = total mean, n = count of records, A = the array of elements

type UserMessageStats struct {
	MsgsLength    []int
	MeanMsgLength int
	MsgsCount     int
	LastMsgTime   int64
	SinceLastMsg  int
}

func (s *UserMessageStats) Reset() {
	// clears the stats
	p := reflect.ValueOf(s).Elem()
	p.Set(reflect.Zero(p.Type()))
}

func (s *UserMessageStats) ControlFlood(isFlood chan bool, job *Job) {
	select {
	case <-isFlood:
		text := fmt.Sprintf(job.app.Features.MessageStatistics.I18n[job.app.Lang].WarnMessage,
			FLOOD_TIME_INTERVAL, FLOOD_MAX_ALLOWED_MSGS, s.MeanMsgLength)

		botEgressReq := &BotEgressSendMessage{
			ChatId:                job.ingressBody.Message.Chat.Id,
			Text:                  text,
			ParseMode:             ParseModeMarkdown,
			DisableWebPagePreview: true,
			DisableNotification:   true,
			ReplyToMessageId:      job.ingressBody.Message.MessageId,
			ReplyMarkup:           &BotForceReply{ForceReply: false, Selective: true},
		}
		botEgressReq.EgressSendToTelegram(job.app)

	case <-time.After(time.Duration(FLOOD_TIME_INTERVAL) * time.Second):
		job.app.Logger.Warn("Resetting user stats")
		s.Reset()
	}
}

func JobMessageStatistics(job *Job) (interface{}, error) {
	if !job.app.Features.MessageStatistics.Enabled || !job.HasMessageContent() {
		return nil, nil
	}

	// 1. get the stats. stats
	stats := UserStatistics[job.ingressBody.Message.From.Id]
	wordsCount := len(strings.Fields(job.ingressBody.Message.Text))
	t := time.Now().Unix()

	if stats == nil {
		// 2.1 init the user stats
		stats = &UserMessageStats{
			MsgsLength:   []int{wordsCount},
			MsgsCount:    1,
			LastMsgTime:  t,
			SinceLastMsg: 0,
		}
	} else {
		// 2.2 update the user stats
		stats.MsgsLength = append(stats.MsgsLength, wordsCount)
		stats.MsgsCount += 1
		stats.MeanMsgLength += ((wordsCount - stats.MeanMsgLength) / stats.MsgsCount) // Ref:formula 1.
		stats.LastMsgTime = t
		stats.SinceLastMsg = int(time.Since(time.Unix(stats.LastMsgTime, 0)).Seconds())
	}

	// 3. update the user stats map
	UserStatistics[job.ingressBody.Message.From.Id] = stats

	// 4.
	isFlood := make(chan bool, 1)
	go stats.ControlFlood(isFlood, job)

	// 5. Detect if user has been flooding for last FLOOD_TIME_INTERVAL seconds
	// add here the condition with the MeanMsgLength within FLOOD_TIME_INTERVAL
	if len(stats.MsgsLength) > FLOOD_MAX_ALLOWED_MSGS {
		job.app.Logger.WithFields(logrus.Fields{
			"userId": job.ingressBody.Message.From.Id,
		}).Warn("User is flooding")

		isFlood <- true
	}

	return stats, nil
}
