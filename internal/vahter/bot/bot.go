package bot

import (
	"log"
	"time"
	"mvdan.cc/xurls"

	redis "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
	router "github.com/novitoll/novitoll_daemon_bot/internal/vahter/router"
)

const (
	duplicateUrlExpiration = time.Duration(14 * 24 * 3600)  // 2 weeks
)

func (br *BotRequest) Process(rh *router.RouteHandler) {
	log.Printf("Username: %s, Chat: %s, Message_Id: %d", br.Message.From.Username, br.Message.Chat.Username, br.Message.Message_Id)

	var urls[] string
	urls = xurls.Relaxed().FindAllString(br.Message.Text, -1)

	if !len(urls) {
		chDuplicationURL := make(chan bool, 1) // buffered?
		go br.CheckDuplicateURL(urls, chDuplicateURL)
		<-chDuplicationURL  // do the action on true result of URL duplication
	}
	
	chAdDetection := make(chan bool, 1) // buffered?	
	go br.CheckForAd(chAdDetection)
	<-chAdDetection // do the action on true result of Ad detection

	// select {

	// }
}

func (br *BotRequest) CheckDuplicateURL(urls []string, ch chan bool) {	
	redisClient := redis.RedisClient()

	pong, err := redisClient.Ping().Result()
	log.Printf(pong, err)

	for _, url := range urls {
		val, err := redisClient.Get(url).Result()
		log.Printf("Found! %s", val)
		if err != nil {
			panic(err)
		}		

		if val != "" {
			log.Printf("Duplicate! %s", url)
			ch <- true
		} else {
			err2 := redisClient.Set(url, br.Message.From, duplicateUrlExpiration).Err()
			ch <- false
			if err2 != nil {
				panic(err)
			}
		}
	}
}
