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

func (br *BotRequest) Process() {
	log.Printf("Username: %s, Chat: %s, Message_Id: %d", br.Message.From.Username, br.Message.Chat.Username, br.Message.Message_Id)

	br.CheckDuplicateURL()
}

func (br *BotRequest) CheckDuplicateURL() {
	var urls[] string
	urls = xurls.Relaxed().FindAllString(br.Message.Text, -1)
	
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
		} else {
			err2 := redisClient.Set(url, br.Message.From, duplicateUrlExpiration).Err()
			if err2 != nil {
				panic(err)
			}
		}
	}
}
