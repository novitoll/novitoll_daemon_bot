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

func (ingressBody *BotIngressRequest) CronJobsStartForChat(app *App) {
	job := &Job{ingressBody, app}

	_, errors := FanOutProcessJobs(job, []ProcessJobFn{
		CronJobNewcomersCount,
		CronJobUserMessageStats,
		CronJobGetChatAdmins,
	})

	for _, e := range errors {
		app.Logger.Fatal(fmt.Sprintf("%v", e))
	}
}

func CronJobGetChatAdmins(job *Job) (interface{}, error) {
	// updates every week on each first ingress request
	var admins []string
	chatId := job.ingressBody.Message.Chat.Id

	adminsReq := &BotEgressGetAdmins{
		ChatId: chatId,
	}
	resp, err := adminsReq.EgressGetAdmins(job.app)
	if err != nil {
		return resp, err
	}
	if len(resp) < 1 {
		job.app.Logger.Warn(fmt.Sprintf("No admins found for chatId: %d", chatId))
		admins = append(admins, fmt.Sprintf("@%s", BDFL))
	} else {
		for _, br := range resp {
			if br.From.Username == TELEGRAM_BOT_USERNAME {
				continue
			}
			admins = append(admins, fmt.Sprintf("@%s", br.From.Username))
		}
	}

	// update the slice of admins
	job.app.ChatAdmins[chatId] = admins

	job.app.Logger.WithFields(logrus.Fields{
		"chatId": chatId,
		"admins": len(admins),
	}).Info("CronJobGetChatAdmins: Completed")

	return nil, nil
}

func CronJobUserMessageStats(job *Job) (interface{}, error) {
	select {
	case <-time.After(time.Duration(EVERY_LAST_SEC_7TH_DAY+5) * time.Second):
		var topKactiveUsers int = 5
		var report []string

		replyTextTpl := job.app.Features.Administration.I18n[job.app.Lang].CronJobUserMsgReport // for short reference

		// we have map of userId:stats
		// we need to put to the ordered slice and sort it by some stats field
		stats := []*UserMessageStats{}
		for _, v := range UserStatistics {
			stats = append(stats, v)
		}
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].AllMsgsCount > stats[i].AllMsgsCount // descending
		})
		// next we select top-K of this sorted slice and do cronjob work
		if len(stats) < topKactiveUsers {
			topKactiveUsers = len(stats)
		}
		for _, userStat := range stats[:topKactiveUsers] {
			report = append(report,
				fmt.Sprintf("\nUser - *%s*, total: %d msgs, avg. msgs length: %d word",
					userStat.Username, userStat.AllMsgsCount, userStat.MeanAllMsgsLength))
		}

		replyText := fmt.Sprintf(replyTextTpl, topKactiveUsers, strings.Join(report, ""))
		resp, err := sendMessage(job, replyText)

		// reset maps
		UserStatistics = make(map[int]*UserMessageStats)
		delete(ChatIds, job.ingressBody.Message.Chat.Id)

		return resp, err
	}
}

func CronJobNewcomersCount(job *Job) (interface{}, error) {
	select {
	case <-time.After(time.Duration(EVERY_LAST_SEC_7TH_DAY) * time.Second):
		var authR, kickR string
		var authN int = len(NewComersAuthVerified)
		var kickN int = len(NewComersKicked)

		replyTextTpl := job.app.Features.Administration.I18n[job.app.Lang].CronJobNewcomersReport // for short reference

		authDiff := utils.CountDiffInPercent(PrevAuth, authN)
		kickDiff := utils.CountDiffInPercent(PrevKick, kickN)

		if authN > 0 {
			authR = fmt.Sprintf("%d(%s)", authN, authDiff)
		} else {
			authR = fmt.Sprintf("%d", authN)
		}

		if kickN > 0 {
			kickR = fmt.Sprintf("%d(%s)", kickN, kickDiff)
		} else {
			kickR = fmt.Sprintf("%d", kickN)
		}

		replyText := fmt.Sprintf(replyTextTpl, authR, kickR)

		resp, err := sendMessage(job, replyText)

		// reset maps
		NewComersAuthVerified = make(map[int]interface{})
		NewComersKicked = make(map[int]interface{})
		// update global counters
		PrevAuth = authN
		PrevKick = kickN

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
		ReplyToMessageId:      0,
		ReplyMarkup:           &BotForceReply{ForceReply: false, Selective: true},
	}
	// notify user about the flood limit
	return botEgressReq.EgressSendToTelegram(job.app)
}
