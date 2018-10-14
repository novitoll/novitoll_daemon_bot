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

func sendHTTP(req *http.Request) bool {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: netTransport,	
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalln("[-] Can not send message to Telegram\n", err)
		return false
	}	
	return true
}

func (reqBody *BotEgressRequest) EgressSendToTelegram(rh *RouteHandler) {
	jsonValue, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(TELEGRAM_URL, TELEGRAM_TOKEN, "sendMessage")
	req, _ := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	sendHTTP(req)
}

func (reqBody *BotEgressKickChatMember) EgressKickChatMember(rh *RouteHandler) bool {
	jsonValue, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(TELEGRAM_URL, TELEGRAM_TOKEN, "kickChatMember")
	req, _ := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	return sendHTTP(req)
}