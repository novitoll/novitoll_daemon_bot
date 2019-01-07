// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/novitoll/novitoll_daemon_bot/internal/utils"
	"github.com/sirupsen/logrus"
)

const (
	EVERY_LAST_SEC_7TH_DAY = 604799
)

var (
	PrevAuth int
	PrevKick int
)

func (ingressBody *BotIngressRequest) StartCronJobsForChat(app *App) {
	job := &Job{ingressBody, app}

	results, errors := FanOutProcessJobs(job, []ProcessJobFn{
		CronJobNewcomersCount,
		CronJobUserMessageStats,
	})

	for _, e := range errors {
		app.Logger.Fatal(fmt.Sprintf("%v", e))
	}

	app.Logger.WithFields(logrus.Fields{
		"completed": len(results),
		"errors":    len(errors),
	}).Info("StartCronJobsForChat: Completed")
}

func CronJobUserMessageStats(job *Job) (interface{}, error) {
	select {
	case <-time.After(time.Duration(EVERY_LAST_SEC_7TH_DAY) * time.Second):
		var topKactiveUsers uint8 = 5
		var report []string

		replyTextTpl := job.app.Features.Administration.I18n[job.app.Lang].CronJobUserMsgReport // for short reference

		// we have map of userId:stats
		// we need to put to the ordered slice and sort it by some stats field
		stats := []*UserMessageStats{}
		for _, v := range UserStatistics {
			stats = append(stats, v)
		}
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].AllMsgsCount > stats[i].AllMsgsCount  // descending
		})
		// next we select top-K of this sorted slice and do cronjob work
		for _, userStat := range stats[:topKactiveUsers] {
			report = append(report, 
				fmt.Sprintf("User - %s. Total: %d msgs, Avg. msgs length: %d",
				 userStat.Username, userStat.AllMsgsCount, userStat.MeanAllMsgsLength))
		}

		replyText := fmt.Sprintf(replyTextTpl, topKactiveUsers, strings.Join(report, ".\n"))
		resp, err := sendMessage(job, replyText)

		// reset maps
		utils.Destruct(&UserStatistics)

		return resp, err
	}
}

func CronJobNewcomersCount(job *Job) (interface{}, error) {
	select {
	case <-time.After(time.Duration(EVERY_LAST_SEC_7TH_DAY) * time.Second):
		replyTextTpl := job.app.Features.Administration.I18n[job.app.Lang].CronJobNewcomersReport // for short reference

		authDiff := utils.CountDiffInPercent(PrevAuth, len(NewComersAuthVerified))
		kickDiff := utils.CountDiffInPercent(PrevKick, len(NewComersKicked))

		replyText := fmt.Sprintf(replyTextTpl, len(NewComersAuthVerified), authDiff, len(NewComersKicked), kickDiff)

		resp, err := sendMessage(job, replyText)

		// reset maps
		NewComersAuthVerified = make(map[int]interface{})
		NewComersKicked = make(map[int]interface{})
		// update global counters
		PrevAuth = len(NewComersAuthVerified)
		PrevKick = len(NewComersKicked)

		return resp, err
	}
}

func sendMessage(job *Job, replyText string) (interface{}, error) {
	botEgressReq := &BotEgressSendMessage{
		ChatId:                job.ingressBody.Message.Chat.Id,
		Text:                  replyText,
		ParseMode:             ParseModeMarkdown,
		DisableWebPagePreview: true,
		DisableNotification:   true,
		ReplyToMessageId:      job.ingressBody.Message.MessageId,
		ReplyMarkup:           &BotForceReply{ForceReply: false, Selective: true},
	}
	// notify user about the flood limit
	return botEgressReq.EgressSendToTelegram(job.app)
}
