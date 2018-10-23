// SPDX-License-Identifier: GPL-2.0
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
	defer response.Body.Close()
	if err != nil {
		log.Fatalln("[-] Can not send message to Telegram\n", err)
		return nil, err
	}

	buf := new(bytes.Buffer)
	parsedBytes, err2 := buf.ReadFrom(response.Body)
	if err2 != nil {
		log.Fatalf("[-] Failed in Telegram response with %d bytes", parsedBytes)
		return nil, err2
	}
	log.Println(buf.String()) // TODO: delete

	// here, for me, it's much easier to try to parse on 2 kind of structs, rather than handle the case with interface{} throughout all calls
	// because Telegram response body values differs with the same key.

	// 1. Try with the usual response body
	var replyMsgBody BotIngressResponse
	err = json.Unmarshal([]byte(buf.String()), &replyMsgBody)
	if err != nil {
		// 2.1 if error, then try with the another one
		var replyMsgBody2 BotIngressResponse2
		err = json.Unmarshal([]byte(buf.String()), &replyMsgBody2)
		if err != nil {
			// 2.2. if error, then return error
			log.Fatalf("[-] Could not parse response body with none of structs")
			return nil, err
		}

		if !replyMsgBody.Ok {
			err = errors.New(fmt.Sprintf("ERROR - %d; %s", replyMsgBody.ErrorCode, replyMsgBody.Description))
			return nil, err
		}
		return nil, err
	} else {
		return &replyMsgBody.Result, err
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
