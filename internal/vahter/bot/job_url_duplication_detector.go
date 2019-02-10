// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"encoding/json"
	"fmt"
	netUrl "net/url"
	"strings"
	"time"

	"github.com/justincampbell/timeago"
	redis "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
	"github.com/sirupsen/logrus"
	"mvdan.cc/xurls"
)

func JobUrlDuplicationDetector(j *Job) (interface{}, error) {
	if !j.app.Features.UrlDuplication.Enabled {
		return false, nil
	}

	urls := xurls.Relaxed().FindAllString(j.req.
		Message.Text, -1)

	if len(urls) == 0 {
		return false, nil
	}

	redisConn := redis.GetRedisConnection()
	defer redisConn.Close()

	// check if user is allowed to post URLs
	// this key expires in Redis in 7 days
	k := fmt.Sprintf("%s-%s", REDIS_USER_VERIFIED, j.req.Message.From.Id)
	user := j.GetFromRedis(redisConn, k)

	if user != nil {
		// new users can not post URLs until 7 days
		n := int(EVERY_LAST_SEC_7TH_DAY + 1/(3600*24))
		botReply := fmt.Sprintf(j.app.Features.NewcomerQuestionnare.
			I18n[j.app.Lang].AuthMessageURLPost, n)

		botReply += fmt.Sprintf("CC: @%s", BDFL)

		go func() {
			select {
			case <-time.After(time.Duration(TIME_TO_DELETE_REPLY_MSG) * time.Second):
				j.DeleteMessage(&j.req.Message)
			}
		}()

		return j.SendMessage(botReply, j.req.Message.MessageId)
	}

	for i, url := range urls {
		// this should be controlled in JobAdDetector
		if strings.Contains(url, "t.me/") {
			continue
		}

		j.app.Logger.Info(fmt.Sprintf("Checking %d/%d URL - %s",
			i+1, len(urls), url))

		if j.app.Features.UrlDuplication.IgnoreHostnames {
			u, err := netUrl.ParseRequestURI(url)
			if err != nil || u.Path == "" {
				continue
			}
		}

		redisKey := strings.ToLower(url)

		jsonStr, _ := redisConn.Get(redisKey).Result()

		if jsonStr != "" {
			j.app.Logger.WithFields(logrus.Fields{
				"url": url,
			}).Warn("This message contains the duplicate URL")

			var duplicatedMsg BotInReqMsg

			json.Unmarshal([]byte(jsonStr), &duplicatedMsg)

			_, err := j.onURLDuplicate(&duplicatedMsg)
			if err != nil {
				return false, err
			}
		} else {
			payload, err := json.Marshal(j.req.Message)

			// should not be the case here
			if err != nil {
				j.app.Logger.Fatal("Can not marshal " +
					"BotInReq.Message from Redis")
				return false, err
			}

			err2 := redisConn.Set(redisKey, payload,
				time.Duration(j.app.Features.UrlDuplication.
					RelevanceTimeout)*time.Second).Err()

			if err2 != nil {
				j.app.Logger.WithFields(logrus.Fields{
					"err": err,
				}).Fatal("[-] Can not put the message to Redis")
				return false, err
			}
		}
	}

	return nil, nil
}

func (j *Job) onURLDuplicate(duplicatedMsg *BotInReqMsg) (
	interface{}, error) {

	j.app.Logger.Info("POST HTTP request on duplicate detection")

	t := time.Since(time.Unix(duplicatedMsg.Date, 0))
	d, _ := time.ParseDuration(t.String())

	botReply := fmt.Sprintf(j.app.Features.UrlDuplication.
		I18n[j.app.Lang].WarnMessage,
		duplicatedMsg.From.Username, timeago.FromDuration(d))

	reply, err := j.SendMessage(botReply, j.req.Message.MessageId)
	if err != nil {
		return false, err
	}

	if reply != nil {
		// cleanup reply messages
		go func() {
			select {
			case <-time.After(time.Duration(TIME_TO_DELETE_REPLY_MSG+10) * time.Second):
				go j.DeleteMessage(reply)
			}
		}()
	}

	return reply, err
}
