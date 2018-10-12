package bot

import (
	"time"
	"log"
	"encoding/json"
	"mvdan.cc/xurls"

	redisClient "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
)

var (
	duplicateUrlExpiration = 14*24*3600*time.Second  // 2 weeks
)

func actionOnURLDuplicate(br *BotIngressRequest, rh *RouteHandler) {
	log.Printf("[+] POST HTTP request on duplicate detection")

	botReplyMessage := "[!] Your message contains duplicate URL. Please dont flood. Last time it was posted:\n"

	botEgressReq := &BotEgressRequest{
		ChatId:					br.Message.Chat.Id,
		Text:					botReplyMessage,
		ParseMode:				ParseModeMarkdown,
		DisableWebPagePreview:	true,
		DisableNotification:	true,
		ReplyToMessageId:		br.Message.MessageId}

	botEgressReq.EgressSendToTelegram(rh)
}

func JobUrlDuplicationDetector(br *BotIngressRequest, rh *RouteHandler) (interface{}, error) {
	redisConn := redisClient.GetRedisConnection() // TODO: improve this using Redis Pool of connections
	defer redisConn.Close()

	urls := xurls.Relaxed.FindAllString(br.Message.Text, -1)
	if len(urls) == 0 {
		return nil, nil
	}

	for i, url := range urls {
		log.Printf("[.] Checking %d/%d URL - %s", i + 1, len(urls), url)
		val, _ := redisConn.Get(url).Result()

		if val != "" {
			log.Printf("[!] This message contains the duplicate URL %s", url)
			go actionOnURLDuplicate(br, rh) // TODO: block with channel waiting for the response from Telegram
		} else {
			fromDataBytes, err := json.Marshal(br.Message)
			if err != nil {
				log.Fatalf("[-] Can not marshal BotIngressRequest.Message from Redis") // should not be the case here
			}

			err2 := redisConn.Set(url, fromDataBytes, duplicateUrlExpiration).Err()
			if err2 != nil {
				panic(err2)
			}
		}
	}

	return nil, nil
}