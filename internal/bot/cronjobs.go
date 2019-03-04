// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	redis_ "github.com/go-redis/redis"
	redis "github.com/novitoll/novitoll_daemon_bot/pkg/redis_client"
	"github.com/novitoll/novitoll_daemon_bot/pkg/utils"
	"github.com/sirupsen/logrus"
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
		j.app.Logger.Warn(fmt.Sprintf("No admins found "+
			"for chatId: %d", chatId))
		admins = append(admins, fmt.Sprintf("@%s", BDFL))
	} else {
		for _, br := range resp {
			if br.From.Username == TELEGRAM_BOT_USERNAME {
				continue
			}
			admins = append(admins, br.From.Username)
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
	case <-time.After(time.Duration(EVERY_LAST_SEC_7TH_DAY+5) *
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

		sort.Slice(stats, func(i, ii int) bool {
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
				fmt.Sprintf("\nUser - *%s*, total: %d msgs, "+
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

		// use 1 redis TCP connection per goroutine
		redisConn := redis.GetRedisConnection()
		defer redisConn.Close()

		// for short reference
		replyTextTpl := j.app.Features.Administration.
			I18n[j.app.Lang].CronJobNewcomersReport

		reports := make([]interface{}, 3)

		for i, s := range []struct {
			redisK  string
			redisKp string
		}{
			{REDIS_USER_VERIFIED, REDIS_USER_PREV_VERIFIED},
			{REDIS_USER_KICKED, REDIS_USER_PREV_KICK},
			{REDIS_USER_LEFT, REDIS_USER_PREV_LEFT},
		} {

			reports[i] = j.__cronUserStats(redisConn, s.redisK, s.redisKp)
		}

		replyText := fmt.Sprintf(replyTextTpl, reports...)

		resp, err := j.SendMessage(replyText, 0)

		delete(ChatIds, j.req.Message.Chat.Id)

		return resp, err
	}
}

func (j *Job) __cronUserStats(redisConn *redis_.Client, redisK string, redisKp string) string {
	// will match all verified users
	// these keys will be expired in Redis in +10 sec
	k := fmt.Sprintf("%s-*", redisK)
	currentUsers := j.GetBatchFromRedis(redisConn, k, 0)

	prevCount := j.GetFromRedis(redisConn, redisKp)
	prevCountI, err := strconv.Atoi(prevCount.(string))
	if err != nil {
		j.app.Logger.Warn("Could not convert string to int")
		return ""
	}

	var N int = len(currentUsers.([]string))

	diff := utils.CountDiffInPercent(prevCountI, N)

	// update counters
	j.SaveInRedis(redisConn, redisKp, N, 0)

	return fmt.Sprintf("%d(%s)", N, diff)
}
