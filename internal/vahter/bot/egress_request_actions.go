package bot

import (
	"fmt"
	"log"
	"bytes"
	"net/http"
	"encoding/json"
)

func sendToTelegram(rh *RouteHandler, reqBody *TelegramRequestBody) {
	jsonValue, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(TELEGRAM_URL, rh.Features.NotificationTarget.Token)
	req, err := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func (br *BotRequest) ActionOnNewcomer(rh *RouteHandler) {
	//
}

func (br *BotRequest) ActionOnURLDuplicate(rh *RouteHandler) {
	log.Printf("[+] POST HTTP request on duplicate detection")

	botReplyMessage := "[!] Your message contains duplicate URL. Please dont flood. Last time it was posted:\n"

	sendToTelegram(rh, &TelegramRequestBody{
		ChatId:			br.Message.Chat.Id,
		Text:			botReplyMessage,
		ParseMode:		ParseModeMarkdown,
		DisableWebPagePreview:	true,
		DisableNotification:	true,
		ReplyToMessageId:	br.Message.MessageId})
}

func (br *BotRequest) ActionOnAdDetection(rh *RouteHandler) {
	//
}
