package bot

import (
	"log"
	"time"
	"mvdan.cc/xurls"

	redis "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
)

const (
	duplicateUrlExpiration = time.Duration(14 * 24 * 3600)  // 2 weeks
)

var redisMock map[string]string

func (br *BotRequest) Process(rh *RouteHandler) {
	log.Printf("Processing message from -- Username: %s, Chat: %s, Message_Id: %d", br.Message.From.Username, br.Message.Chat.Username, br.Message.Message_Id)

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

func (br *BotRequest) CheckDuplicateURL(urls []string, ch chan bool) {	
	redisClient := redis.RedisClient()

	for _, url := range urls {
		val, err := redisClient.Get(url).Result()
		if err != nil {
			panic(err)
		}		

		if val != "" {
			log.Printf("Duplicate! %s", url)
			ch <- true
		} else {
			if err2 := redisClient.Set(url, br.Message.From, duplicateUrlExpiration).Err();err2 != nil {
				panic(err)
			}
			redisMock[url] = br.Message.From.Username
			ch <- false
		}
	}
}
