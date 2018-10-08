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
	log.Printf("[.] Processing message from -- Username: %s, Chat: %s, Message_Id: %d", br.Message.From.Username, br.Message.Chat.Username, br.Message.MessageId)

	chNewcomerDetection := make(chan bool) // buffered?
	chDuplicationURL := make(chan bool) // buffered?
	chAdDetection := make(chan bool) // buffered?

	if rh.Features.NewcomerQuestionnare.Enabled {
		go br.CheckNewcomer(chNewcomerDetection)
	}

	if rh.Features.UrlDuplication.Enabled {
		redisConn := redisClient.GetRedisConnection()
		go br.CheckDuplicateURL(chDuplicationURL, redisConn)
	}

	if rh.Features.AdDetection.Enabled {
		go br.CheckForAd(chAdDetection)
	}

	// TODO: need to investigate/benchmark this carefully, buffered chan might be better, sync.Mutex?
	for {
		select {
			case isNewcomer := <-chNewcomerDetection:  // do the action on true result of Newcomer detection
				if isNewcomer {
					go br.ActionOnNewcomer(rh)
				}
			case isDuplicateURL := <-chDuplicationURL:  // do the action on true result of URL duplication
				if isDuplicateURL {
					go br.ActionOnURLDuplicate(rh)
				}
			case isAd := <-chAdDetection: // do the action on true result of Ad detection
				if isAd {
					go br.ActionOnAdDetection(rh)
				}
			default:
				br.CountStatistics(rh)
		}
	}
}

func (br *BotRequest) CheckDuplicateURL(ch chan bool, redisConn *redis.Client) {
	defer redisConn.Close()

	urls := xurls.Relaxed.FindAllString(br.Message.Text, -1)
	if len(urls) == 0 {
		ch <- false
		return
	}

	for i, url := range urls {
		log.Printf("[.] Checking %d/%d URL - %s", i + 1, len(urls), url)
		val, _ := redisConn.Get(url).Result()

		if val != "" {
			// var duplicateBrMsg BotRequest
			// json.Unmarshal(val, &duplicateBrMsg)
			// log.Printf("[!] This message contains the duplicate URL %s. \nFrom: %s, Date: %d, Chat: %s", url, duplicateBrMsg.From.Username, duplicateBrMsg.Date, duplicateBrMsg.Chat.Username)
			log.Printf("[!] This message contains the duplicate URL %s", url)
			ch <- true
			return
		} else {
			fromDataBytes, err := json.Marshal(br.Message)
			if err != nil {
				log.Fatalf("[-] Can not marshal BotRequest.Message from Redis") // should not be the case here
				ch <- false
				return
			}

			err2 := redisConn.Set(url, fromDataBytes, duplicateUrlExpiration).Err()
			if err2 != nil {
				panic(err2)
			}
			ch <- false
		}
	}
}

func (br *BotRequest) CheckNewcomer(ch chan bool) {
	// put the newcomer ID to the Redis for 48h expiration
	// before expiration notify admins that newcomers have not said a word
	if br.Message.NewComer.Username != "" {

	}
}
