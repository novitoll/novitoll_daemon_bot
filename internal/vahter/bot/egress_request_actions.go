// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func sendHTTP(req *http.Request, app *App) (*bytes.Buffer, error) {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	app.Logger.Info("POST HTTP egress to Telegram")

	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	response, err := client.Do(req)
	if err != nil {
		app.Logger.WithFields(logrus.Fields{"err": err}).Fatal("Can not send message to Telegram")
		return nil, err
	}
	defer response.Body.Close()

	buf := new(bytes.Buffer)
	parsedBytes, err2 := buf.ReadFrom(response.Body)
	if err2 != nil {
		app.Logger.WithFields(logrus.Fields{"parsedBytes": parsedBytes}).Fatal("Failed in Telegram response")
		return nil, err2
	}
	app.Logger.Info(buf.String()) // TODO: delete

	return buf, err2
}

func parseBody(req *http.Request, app *App) (*BotIngressRequestMessage, error) {
	buf, err := sendHTTP(req, app)
	if err != nil {
		return nil, err
	}

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
			app.Logger.Fatal("Could not parse response body with none of structs")
			return nil, err
		}

		if !replyMsgBody.Ok {
			err = errors.New(fmt.Sprintf("ERROR - %d; %s", replyMsgBody.ErrorCode, replyMsgBody.Description))
			return nil, err
		}
		return nil, err
	}
	return &replyMsgBody.Result, err
}

func (app *App) EgressRequest(reqBody interface{}, method string) (*BotIngressRequestMessage, error) {
	jsonValue, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(TELEGRAM_URL, TELEGRAM_TOKEN, method)
	req, _ := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	return parseBody(req, app)
}

func (reqBody *BotEgressSendMessage) EgressSendToTelegram(app *App) (*BotIngressRequestMessage, error) {
	return app.EgressRequest(reqBody, "sendMessage")
}

func (reqBody *BotEgressKickChatMember) EgressKickChatMember(app *App) (*BotIngressRequestMessage, error) {
	return app.EgressRequest(reqBody, "kickChatMember")
}

func (reqBody *BotEgressDeleteMessage) EgressDeleteMessage(app *App) (*BotIngressRequestMessage, error) {
	return app.EgressRequest(reqBody, "deleteMessage")
}

func (reqBody *BotEgressGetAdmins) EgressGetAdmins(app *App) ([]*BotIngressRequestMessage, error) {
	jsonValue, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(TELEGRAM_URL, TELEGRAM_TOKEN, "getChatAdministrators")
	req, _ := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	buf, err := sendHTTP(req, app)
	if err != nil {
		return nil, err
	}

	var replyMsgBody BotIngressResponseMult
	err = json.Unmarshal([]byte(buf.String()), &replyMsgBody)
	if err != nil {
		return nil, err
	}
	return replyMsgBody.Result, err
}

func (j *Job) DeleteMessage(response *BotIngressRequestMessage) (interface{}, error) {
	j.app.Logger.WithFields(logrus.Fields{
		"id": response.MessageId,
	}).Info("Deleting a reply message")

	botEgressReq := &BotEgressDeleteMessage{
		ChatId:    response.Chat.Id,
		MessageId: response.MessageId,
	}
	_, err := botEgressReq.EgressDeleteMessage(j.app)
	if err != nil {
		return false, err
	}
	return true, nil
}
