package bot

import (
	"log"
	"regexp"
)

var (
	regexp_http = regexp.MustCompile(`(http|https|ftp|ftps)\:\/\/[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,3}(\/\S*)?`)
)

func (br *BotRequest) Process() {
	log.Printf("Username: %s, Chat: %s, Message_Id: %d", br.Message.From.Username, br.Message.Chat.Username, br.Message.Message_Id)
}

func is_hyperlink(text *string) bool{
	return regexp_http.MatchString(*text)
}
