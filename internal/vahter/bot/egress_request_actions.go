package bot

import (
	"fmt"
	"log"
	"bytes"
	"time"
	"net"
	"net/http"
	"encoding/json"
)

func sendHTTP(req *http.Request) (*BotIngressRequest, error) {
	var replyMsgBody BotIngressRequest

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	
	// TODO remove
	log.Println("[!] POST HTTP egress to Telegram")

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: netTransport,
	}
	response, err := client.Do(req)
	if err != nil {
		log.Fatalln("[-] Can not send message to Telegram\n", err)
		return nil, err
	}
	
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	json.Unmarshal([]byte(buf.String()), &replyMsgBody)

	defer response.Body.Close()
	
	return &replyMsgBody, err
}

// TODO: Factory? these functions are similar, difference is request body and Telegram command

func (reqBody *BotEgressSendMessage) EgressSendToTelegram(app *App) (*BotIngressRequest, error) {
	jsonValue, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(TELEGRAM_URL, TELEGRAM_TOKEN, "sendMessage")
	req, _ := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	return sendHTTP(req)
}

func (reqBody *BotEgressKickChatMember) EgressKickChatMember(app *App) (*BotIngressRequest, error) {
	jsonValue, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(TELEGRAM_URL, TELEGRAM_TOKEN, "kickChatMember")
	req, _ := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	return sendHTTP(req)
}

func (reqBody *BotEgressDeleteMessage) EgressDeleteMessage(app *App) (*BotIngressRequest, error) {
	jsonValue, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(TELEGRAM_URL, TELEGRAM_TOKEN, "deleteMessage")
	req, _ := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	return sendHTTP(req)
}