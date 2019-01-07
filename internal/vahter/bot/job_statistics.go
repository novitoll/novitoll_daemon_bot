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
)

var (
	// Map to store user message statistics.
	// Data in the map is cleaned up when the CronJob executes (every last second of 7th day)
	UserStatistics = make(map[int]*UserMessageStats)
)

// formula 1. (Incremental average) M_n = M_(n-1) + ((A_n - M_(n-1)) / n), where M_n = total mean, n = count of records, A = the array of elements

type UserMessageStats struct {
	FloodMsgsLength   []int
	AllMsgsCount      int
	LastMsgTime       int64
	SinceLastMsg      int
	MeanAllMsgsLength int
	Dates 			  []int64
	Username 		  string
}

func JobMessageStatistics(job *Job) (interface{}, error) {
	if !job.app.Features.MessageStatistics.Enabled || !job.HasMessageContent() {
		return nil, nil
	}

	// 1. get the stats. stats
	stats := UserStatistics[job.ingressBody.Message.From.Id]
	wordsCount := len(strings.Fields(job.ingressBody.Message.Text))
	t0 := time.Now().Unix()

	if stats == nil {
		// 2.1 init the user stats
		stats = &UserMessageStats{
			FloodMsgsLength:   []int{wordsCount},
			AllMsgsCount:      0,
			LastMsgTime:       t0,
			SinceLastMsg:      0,
			MeanAllMsgsLength: 0,
			Username: job.ingressBody.Message.From.Username,
		}
	} else {
		// 2.2 update the user stats
		stats.Dates = append(stats.Dates, job.ingressBody.Message.Date)
		stats.FloodMsgsLength = append(stats.FloodMsgsLength, wordsCount)
		stats.AllMsgsCount += 1
		stats.MeanAllMsgsLength += ((wordsCount - stats.MeanAllMsgsLength) / stats.AllMsgsCount) // Ref:formula 1.
		stats.SinceLastMsg = int(time.Since(time.Unix(stats.LastMsgTime, 0)).Seconds())
		stats.LastMsgTime = t0
	}

	// 3. Detect if user has been flooding for last TIME_INTERVAL seconds
	// add here the condition with the MeanMsgLength within TIME_INTERVAL
	err := floodDetection(job, stats)

	return stats, err
}

func floodDetection(job *Job, stats *UserMessageStats) error {
	var isFlood bool
	var replyText []string
	replyTextTpl := job.app.Features.MessageStatistics.I18n[job.app.Lang] // for short reference

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
		isFlood = true
		replyText = append(replyText, fmt.Sprintf(replyTextTpl.WarnMessageTooFreq,
			FLOOD_TIME_INTERVAL, FLOOD_MAX_ALLOWED_MSGS))
	}

	if allWordsCount := utils.SumSliceInt(stats.FloodMsgsLength); allWordsCount >= FLOOD_MAX_ALLOWED_WORDS {
		isFlood = true
		replyText = append(replyText, fmt.Sprintf(replyTextTpl.WarnMessageTooLong, FLOOD_MAX_ALLOWED_WORDS))

		// notify admins
		adminsToNotify := strings.Join(job.app.Features.Administration.Admins, ", ")
		replyText = append(replyText, fmt.Sprintf(". CC: %s", adminsToNotify))
	}

	if isFlood {
		botEgressReq := &BotEgressSendMessage{
			ChatId:                job.ingressBody.Message.Chat.Id,
			Text:                  strings.Join(replyText, ". "),
			ParseMode:             ParseModeMarkdown,
			DisableWebPagePreview: true,
			DisableNotification:   true,
			ReplyToMessageId:      job.ingressBody.Message.MessageId,
			ReplyMarkup:           &BotForceReply{ForceReply: false, Selective: true},
		}
		// notify user about the flood limit
		replyMsgBody, err := botEgressReq.EgressSendToTelegram(job.app)
		if err != nil {
			return err
		}

		if replyMsgBody != nil {
			// cleanup reply messages
			go func() {
				select {
				case <-time.After(time.Duration(TIME_TO_DELETE_REPLY_MSG+10) * time.Second):
					job.DeleteMessage(replyMsgBody)
				}
			}()
		}
	}

	if stats.SinceLastMsg > FLOOD_TIME_INTERVAL {
		stats.FloodMsgsLength = []int{}
	}

	// 4. update the user stats map
	UserStatistics[job.ingressBody.Message.From.Id] = stats
	return nil
}
