// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	FLOOD_TIME_INTERVAL    = 10
	FLOOD_MAX_ALLOWED_MSGS = 3
)

var (
	UserStatistics = make(map[int]*UserMessageStats)
)

// formula 1. (Incremental average) M_n = M_(n-1) + ((A_n - M_(n-1)) / n), where M_n = total mean, n = count of records, A = the array of elements

type UserMessageStats struct {
	FloodMsgsLength   []int
	AllMsgsCount      int
	LastMsgTime       int64
	SinceLastMsg      int
	MeanAllMsgsLength int
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
			FloodMsgsLength:   []int{wordsCount},
			AllMsgsCount:      0,
			LastMsgTime:       t,
			SinceLastMsg:      0,
			MeanAllMsgsLength: 0,
		}
	} else {
		// 2.2 update the user stats
		stats.FloodMsgsLength = append(stats.FloodMsgsLength, wordsCount)
		stats.AllMsgsCount += 1
		stats.MeanAllMsgsLength += ((wordsCount - stats.MeanAllMsgsLength) / stats.AllMsgsCount) // Ref:formula 1.
		stats.SinceLastMsg = int(time.Since(time.Unix(stats.LastMsgTime, 0)).Seconds())
		stats.LastMsgTime = t
	}

	// 3. Detect if user has been ng for last TIME_INTERVAL seconds
	// add here the condition with the MeanMsgLength within TIME_INTERVAL

	// 5 < 10 && 6 >= 5 -- flood
	// 20 > 10 -- not flood
	// 5 > 10 && 4 <= 5

	if stats.SinceLastMsg <= FLOOD_TIME_INTERVAL && len(stats.FloodMsgsLength) >= FLOOD_MAX_ALLOWED_MSGS {
		job.app.Logger.WithFields(logrus.Fields{
			"userId": job.ingressBody.Message.From.Id,
		}).Warn("User is flooding")

		job.app.Logger.WithFields(logrus.Fields{
			"AllMsgsCount":      stats.AllMsgsCount,
			"LastMsgTime":       stats.LastMsgTime,
			"SinceLastMsg":      stats.SinceLastMsg,
			"MeanAllMsgsLength": stats.MeanAllMsgsLength,
		}).Warn("Resetting user stats")

		stats.FloodMsgsLength = []int{}

		text := fmt.Sprintf(job.app.Features.MessageStatistics.I18n[job.app.Lang].WarnMessage,
			FLOOD_TIME_INTERVAL, FLOOD_MAX_ALLOWED_MSGS)

		botEgressReq := &BotEgressSendMessage{
			ChatId:                job.ingressBody.Message.Chat.Id,
			Text:                  text,
			ParseMode:             ParseModeMarkdown,
			DisableWebPagePreview: true,
			DisableNotification:   true,
			ReplyToMessageId:      job.ingressBody.Message.MessageId,
			ReplyMarkup:           &BotForceReply{ForceReply: false, Selective: true},
		}
		return botEgressReq.EgressSendToTelegram(job.app)
	}

	if stats.SinceLastMsg > FLOOD_TIME_INTERVAL {
		stats.FloodMsgsLength = []int{}
	}

	// 4. update the user stats map
	UserStatistics[job.ingressBody.Message.From.Id] = stats

	return stats, nil
}
