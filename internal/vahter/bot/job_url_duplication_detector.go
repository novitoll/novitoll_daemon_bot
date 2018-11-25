// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"encoding/json"
	"fmt"
	netUrl "net/url"
	"time"

	"github.com/justincampbell/timeago"
	redisClient "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
	"github.com/sirupsen/logrus"
	"mvdan.cc/xurls"
)

func JobUrlDuplicationDetector(j *Job) (interface{}, error) {
	if !j.app.Features.UrlDuplication.Enabled {
		return false, nil
	}

	urls := xurls.Relaxed().FindAllString(j.ingressBody.Message.Text, -1)
	if len(urls) == 0 {
		return false, nil
	}

	redisConn := redisClient.GetRedisConnection() // TODO: improve this using Redis Pool of connections
	defer redisConn.Close()

	for i, url := range urls {
		j.app.Logger.Info(fmt.Sprintf("Checking %d/%d URL - %s", i+1, len(urls), url))

		if j.app.Features.UrlDuplication.IgnoreHostnames {
			u, err := netUrl.ParseRequestURI(url)
			if err != nil || u.Path == "" {
				continue
			}
		}

		// Redis key is constructed via channel Id in order to let the single bot binary operate on multiple chats
		redisKey := fmt.Sprintf("%d-%s", j.ingressBody.Message.Chat.Id, url)

		jsonStr, _ := redisConn.Get(redisKey).Result()

		if jsonStr != "" {
			j.app.Logger.WithFields(logrus.Fields{
				"url": url,
			}).Warn("This message contains the duplicate URL")

			var duplicatedMsg BotIngressRequestMessage
			json.Unmarshal([]byte(jsonStr), &duplicatedMsg)
			_, err := j.actionOnURLDuplicate(&duplicatedMsg)
			if err != nil {
				return false, err
			}
		} else {
			fromDataBytes, err := json.Marshal(j.ingressBody.Message)
			if err != nil {
				j.app.Logger.Fatal("Can not marshal BotIngressRequest.Message from Redis") // should not be the case here
				return false, err
			}

			err2 := redisConn.Set(redisKey, fromDataBytes, time.Duration(j.app.Features.UrlDuplication.RelevanceTimeout)*time.Second).Err()
			if err2 != nil {
				j.app.Logger.WithFields(logrus.Fields{
					"err": err,
				}).Fatal("[-] Can not put the message to Redis")
				return false, err
			}
		}
	}

	return true, nil
}

func (j *Job) actionOnURLDuplicate(duplicatedMsg *BotIngressRequestMessage) (interface{}, error) {
	j.app.Logger.Info("POST HTTP request on duplicate detection")

	t := time.Since(time.Unix(duplicatedMsg.Date, 0))
	d, _ := time.ParseDuration(t.String())

	botReplyMessage := fmt.Sprintf(j.app.Features.UrlDuplication.I18n[j.app.Lang].WarnMessage,
		duplicatedMsg.From.Username, timeago.FromDuration(d))

	reply := &BotForceReply{ForceReply: true, Selective: true}

	botEgressReq := &BotEgressSendMessage{
		ChatId:                j.ingressBody.Message.Chat.Id,
		Text:                  botReplyMessage,
		ParseMode:             ParseModeMarkdown,
		DisableWebPagePreview: true,
		DisableNotification:   true,
		ReplyToMessageId:      j.ingressBody.Message.MessageId,
		ReplyMarkup:           reply,
	}

	replyMsgBody, err := botEgressReq.EgressSendToTelegram(j.app)
	if err != nil {
		return false, err
	}

	if replyMsgBody != nil {
		// cleanup reply messages
		go func() {
			select {
			case <-time.After(time.Duration(TIME_TO_DELETE_REPLY_MSG+10) * time.Second):
				go j.DeleteMessage(replyMsgBody)
			}
		}()
	}

	return replyMsgBody, err
}
