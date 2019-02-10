// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/novitoll/novitoll_daemon_bot/internal/utils"
	// redis "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
	"github.com/sirupsen/logrus"
)

var (
	PrevAuth int
	PrevKick int
)

const (
	EVERY_LAST_SEC_7TH_DAY = 604799
)

func (req *BotInReq) CronSchedule(app *App) {
	j := &Job{req, app}

	_, errors := FanOutProcessJobs(j, []ProcessJobFn{
		CronUserStats,
		// CronChatMsgStats, BUG: fix
		CronGetChatAdmins,
	})

	for _, e := range errors {
		app.Logger.Fatal(fmt.Sprintf("%v", e))
	}
}

func CronGetChatAdmins(j *Job) (interface{}, error) {
	// updates every week on each first ingress request
	var admins []string
	chatId := j.req.Message.Chat.Id

	adminsReq := &BotGetAdmins{
		ChatId: chatId,
	}

	resp, err := adminsReq.GetAdmins(j.app)
	
	if err != nil {
		return resp, err
	}

	if len(resp) < 1 {
		j.app.Logger.Warn(fmt.Sprintf("No admins found " +
			"for chatId: %d", chatId))
		admins = append(admins, fmt.Sprintf("@%s", BDFL))
	} else {
		for _, br := range resp {
			if br.From.Username == TELEGRAM_BOT_USERNAME {
				continue
			}
			admins = append(admins, fmt.Sprintf("@%s",
				br.From.Username))
		}
	}

	// update the slice of admins
	j.app.ChatAdmins[chatId] = admins

	j.app.Logger.WithFields(logrus.Fields{
		"chatId": chatId,
		"admins": len(admins),
	}).Info("CronGetChatAdmins: Completed")

	return nil, nil
}

func CronChatMsgStats(j *Job) (interface{}, error) {
	select {
	case <-time.After(time.Duration(EVERY_LAST_SEC_7TH_DAY + 5) *
		time.Second):

		var topKactiveUsers int = 5
		var report []string

		// for short reference
		replyTextTpl := j.app.Features.Administration.
			I18n[j.app.Lang].CronJobUserMsgReport

		// we have map of userId:stats
		// we need to put to the ordered slice and sort it by 
		// some stats field
		stats := []*UserMessageStats{}
		for _, v := range UserStatistics {
			stats = append(stats, v)
		}

		sort.Slice(stats, func(i, j int) bool {
			// descending sort
			return stats[i].AllMsgsCount > stats[i].AllMsgsCount
		})

		// next we select top-K of this sorted slice and
		// do cronj work
		if len(stats) < topKactiveUsers {
			topKactiveUsers = len(stats)
		}

		for _, userStat := range stats[:topKactiveUsers] {
			report = append(report,
				fmt.Sprintf("\nUser - *%s*, total: %d msgs, " +
					"avg. msgs length: %d word",
					userStat.Username, userStat.AllMsgsCount, 
					userStat.MeanAllMsgsLength))
		}

		replyText := fmt.Sprintf(replyTextTpl, topKactiveUsers,
			strings.Join(report, ""))
		resp, err := j.SendMessage(replyText, 0)

		// reset maps
		UserStatistics = make(map[int]*UserMessageStats)
		delete(ChatIds, j.req.Message.Chat.Id)

		return resp, err
	}
}

func CronUserStats(j *Job) (interface{}, error) {
	select {
	case <-time.After(time.Duration(EVERY_LAST_SEC_7TH_DAY) * 
		time.Second):

		var authR, kickR string
		var authN int = len(NewComersAuthVerified)
		var kickN int = len(NewComersKicked)

		// for short reference
		replyTextTpl := j.app.Features.Administration.
			I18n[j.app.Lang].CronJobNewcomersReport

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

		resp, err := j.SendMessage(replyText, 0)

		// reset maps
		NewComersAuthVerified = make(map[int]interface{})
		NewComersKicked = make(map[int]interface{})
		// update global counters
		PrevAuth = authN
		PrevKick = kickN

		return resp, err
	}
}