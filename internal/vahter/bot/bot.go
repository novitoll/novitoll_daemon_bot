package bot

import (
	"log"
	"time"
	"mvdan.cc/xurls"
	"encoding/json"
	"github.com/go-redis/redis"

	redisClient "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
)

const (
	duplicateUrlExpiration = 14*24*3600*time.Second  // 2 weeks
)

func (br *BotRequest) Process(rh *RouteHandler) {
	log.Printf("[.] Processing message from -- Username: %s, Chat: %s, Message_Id: %d", br.Message.From.Username, br.Message.Chat.Username, br.Message.Message_Id)

	if rh.Features.UrlDuplication.Enabled {
		urls := xurls.Relaxed.FindAllString(br.Message.Text, -1)

		if len(urls) != 0 {
			chDuplicationURL := make(chan bool) // buffered?
			go br.CheckDuplicateURL(urls, chDuplicationURL)
			<-chDuplicationURL  // do the action on true result of URL duplication
		}

	} else if rh.Features.NewcomerQuestionnare.Enabled {
		chAdDetection := make(chan bool) // buffered?
		go br.CheckForAd(chAdDetection)
		<-chAdDetection // do the action on true result of Ad detection
	}

	// select {

	// }
}

func getRedisConnection() *redis.Client {
	rc := redisClient.RedisClient{nil}
	rc.Connect()
	return rc.Conn
}

func (br *BotRequest) CheckDuplicateURL(urls []string, ch chan bool) {
	for _, url := range urls {
		log.Printf("[.] Checking URL - %s", url)
		val, _ := getRedisConnection().Get(url).Result()

		log.Printf("[.] Got the result from Redis %v", val)

		if val != "" {
			log.Printf("Duplicate! %s", url)
			ch <- true
		} else {
			fromDataBytes, err := json.Marshal(br.Message.From)
			if err != nil {
				log.Fatalf("[-] Can not marshal Message.From BotRequest")
				ch <- false
				return
			}

			log.Printf("%t", url)

			err2 := getRedisConnection().Set(url, fromDataBytes, duplicateUrlExpiration).Err()
			if err2 != nil {
				panic(err2)
			}
			ch <- false
		}
	}
}
