package main

import (
	"log"
	"os"
	"regexp"
	"gopkg.in/telegram-bot-api.v4"
	"strings"
)

var (
	regexp_http = regexp.MustCompile(`(http|https|ftp|ftps)\:\/\/[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,3}(\/\S*)?`)
)

type User struct {
	username string
}

type MessageWithURL map[string]*User

func main() {
	token := os.Args[1]

	msgsWithURL := make(MessageWithURL)
	var replyMsg strings.Builder

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		ptext := &update.Message.Text
		pusername := &update.Message.From.UserName

		// deal only with hyperlinks for now
		if is_hyperlink(ptext) {
			log.Printf("[%s] %s", *pusername, *ptext)

			if _, ok := msgsWithURL[*ptext]; ok {
				replyMsg.WriteString("Duplicated URL: ")
				replyMsg.WriteString(*ptext)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyMsg.String())
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)

				replyMsg.Reset()
				continue
			}

			msgsWithURL[*ptext] = &User{username: *pusername}
		}
	}
}


func is_hyperlink(text *string) bool{
	return regexp_http.MatchString(*text)
}
