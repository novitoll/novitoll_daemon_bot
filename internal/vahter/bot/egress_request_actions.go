package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

func sendHTTP(req *http.Request) (*BotIngressRequestMessage, error) {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	log.Println("[.] POST HTTP egress to Telegram")

	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	response, err := client.Do(req)
	if err != nil {
		log.Fatalln("[-] Can not send message to Telegram\n", err)
		return nil, err
	}

	if response.Body != nil {
		var replyMsgBody BotIngressResponse
		buf := new(bytes.Buffer)
		buf.ReadFrom(response.Body)
		json.Unmarshal([]byte(buf.String()), &replyMsgBody)
		defer response.Body.Close()

		if !replyMsgBody.Ok {
			err = errors.New(fmt.Sprintf("ERROR - %d; %s", replyMsgBody.ErrorCode, replyMsgBody.Description))
			return nil, err
		} else {
			return &replyMsgBody.Result, err
		}
	} else {
		return nil, nil
	}
}

// TODO: Factory? these functions are similar, difference is request body and Telegram command

func (reqBody *BotEgressSendMessage) EgressSendToTelegram(app *App) (*BotIngressRequestMessage, error) {
	jsonValue, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(TELEGRAM_URL, TELEGRAM_TOKEN, "sendMessage")
	req, _ := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	return sendHTTP(req)
}

func (reqBody *BotEgressKickChatMember) EgressKickChatMember(app *App) (*BotIngressRequestMessage, error) {
	jsonValue, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(TELEGRAM_URL, TELEGRAM_TOKEN, "kickChatMember")
	req, _ := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	return sendHTTP(req)
}

func (reqBody *BotEgressDeleteMessage) EgressDeleteMessage(app *App) (*BotIngressRequestMessage, error) {
	jsonValue, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(TELEGRAM_URL, TELEGRAM_TOKEN, "deleteMessage")
	req, _ := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	return sendHTTP(req)
}
