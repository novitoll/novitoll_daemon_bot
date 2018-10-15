package bot

import (
	"fmt"
	"time"
	"log"
	"encoding/json"
	"mvdan.cc/xurls"
	"github.com/justincampbell/timeago"
	
	netUrl "net/url"
	redisClient "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
)

func (j *Job) actionOnURLDuplicate(duplicatedMsg *BotIngressRequestMessage) {
	log.Printf("[+] POST HTTP request on duplicate detection")

	t := time.Since(time.Unix(duplicatedMsg.Date, 0))
	d, _ := time.ParseDuration(t.String())

	botReplyMessage := "Your message contains duplicate URL. Please dont flood.\n"
	botReplyMessage += fmt.Sprintf("Last time it was posted from @%s %s ago. #novitollurl", duplicatedMsg.From.Username, timeago.FromDuration(d))

	botEgressReq := &BotEgressSendMessage{
		ChatId:					j.ingressBody.Message.Chat.Id,
		Text:					botReplyMessage,
		ParseMode:				ParseModeMarkdown,
		DisableWebPagePreview:	true,
		DisableNotification:	true,
		ReplyToMessageId:		j.ingressBody.Message.MessageId,
		ReplyMarkup:			&BotForceReply{ForceReply: true, Selective: true}}

	botEgressReq.EgressSendToTelegram(j.app)
}

func JobUrlDuplicationDetector(j *Job) (bool, error) {
	if !j.app.Features.UrlDuplication.Enabled {
		return false, nil
	}
	redisConn := redisClient.GetRedisConnection() // TODO: improve this using Redis Pool of connections
	defer redisConn.Close()

	urls := xurls.Relaxed.FindAllString(j.ingressBody.Message.Text, -1)
	if len(urls) == 0 {
		return false, nil
	}

	for i, url := range urls {
		log.Printf("[.] Checking %d/%d URL - %s", i + 1, len(urls), url)

		if j.app.Features.UrlDuplication.IgnoreHostnames {
			u, err := netUrl.ParseRequestURI(url)
			if err != nil || u.Path == "" {
				log.Printf("[.] Skipping a hostname URL")
				continue
			}
		}

		redisKey := fmt.Sprintf("%d-%s", j.ingressBody.Message.Chat.Id, url)

		jsonStr, _ := redisConn.Get(redisKey).Result()

		if jsonStr != "" {
			log.Printf("[!] This message contains the duplicate URL %s", url)
			var duplicatedMsg BotIngressRequestMessage
			json.Unmarshal([]byte(jsonStr), &duplicatedMsg)
			j.actionOnURLDuplicate(&duplicatedMsg)
		} else {
			fromDataBytes, err := json.Marshal(j.ingressBody.Message)
			if err != nil {
				log.Fatalf("[-] Can not marshal BotIngressRequest.Message from Redis") // should not be the case here
			}			

			err2 := redisConn.Set(redisKey, fromDataBytes, time.Duration(j.app.Features.UrlDuplication.RelevanceTimeout) * time.Second).Err()
			if err2 != nil {
				log.Fatalln("[-] Can not put the message to Redis\n", err2)
				// TODO: notify admin
			}
		}
	}

	return true, nil
}