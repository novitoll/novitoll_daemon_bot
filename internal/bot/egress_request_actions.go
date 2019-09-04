// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
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

	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	resp, err := client.Do(req)
	if err != nil {
		app.Logger.WithFields(logrus.Fields{"err": err}).
			Warn("Can not send message to Telegram")
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	parsedBytes, err2 := buf.ReadFrom(resp.Body)
	if err2 != nil {
		app.Logger.WithFields(logrus.Fields{
			"parsedBytes": parsedBytes,
		}).Warn("Failed in Telegram resp")
		return nil, err2
	}

	return buf, err2
}

func parseBody(req *http.Request, app *App) (*BotInReqMsg, error) {
	buf, err := sendHTTP(req, app)
	if err != nil {
		return nil, err
	}

	// here, for me, it's much easier to try to
	// parse on 2 kind of structs, rather than
	// handle the case with interface{}
	// throughout all calls because Telegram
	// resp body values differs with the same key.

	// 1. Try with the usual resp body
	var reply BotInResp
	err = json.Unmarshal([]byte(buf.String()), &reply)

	if err != nil {

		// 2.1 if error, then try with the another one
		var reply2 BotInResp2
		err = json.Unmarshal([]byte(buf.String()), &reply2)

		if err != nil {
			// 2.2. if error, then return error
			app.Logger.Fatal("Could not parse resp body with none of structs")
			return nil, err
		}

		if !reply.Ok {
			err = errors.New(fmt.Sprintf("ERROR - %d; %s",
				reply.ErrorCode, reply.Description))
			return nil, err
		}
		return nil, err
	}
	return &reply.Result, err
}

func (app *App) SendToTelegram(body interface{}, method string) (*BotInReqMsg, error) {
	jsonValue, _ := json.Marshal(body)
	url := fmt.Sprintf(TELEGRAM_URL, TELEGRAM_TOKEN, method)
	req, _ := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	return parseBody(req, app)
}

// These are wrapper functions of each struct (which is JSON req. body of Telegram API).
// We might use here interface but since each API has separate behavior,
// let's leave it as it's for now.

func (body *BotSendMsg) SendMsg(app *App) (*BotInReqMsg, error) {
	// Telegram parses underscore as the markdown
	body.Text = strings.Replace(body.Text, "_", "\\_", -1)
	return app.SendToTelegram(body, "sendMessage")
}

func (body *BotKickChatMember) KickChatMember(app *App) (*BotInReqMsg, error) {
	return app.SendToTelegram(body, "kickChatMember")
}

func (body *BotDeleteMsg) DeleteMsg(app *App) (*BotInReqMsg, error) {
	return app.SendToTelegram(body, "deleteMessage")
}

func (body *BotAnswerCallbackQuery) AnswerCallbackQuery(app *App) (*BotInReqMsg, error) {
	return app.SendToTelegram(body, "answerCallbackQuery")
}

// This is the same wrapper as above struct functions,
// but its response and request parsing are different,
// so have to duplicate the code without ready func. wrappers
func (body *BotGetAdmins) GetAdmins(app *App) ([]*BotInReqMsg, error) {
	jsonValue, _ := json.Marshal(body)
	url := fmt.Sprintf(TELEGRAM_URL, TELEGRAM_TOKEN, "getChatAdministrators")

	req, _ := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	buf, err := sendHTTP(req, app)
	if err != nil {
		return nil, err
	}

	var reply BotInRespMult
	var replyStr string

	// dirty hack because Telegram resp keys are different
	replyStr = strings.Replace(buf.String(),
		"{\"user", "{\"from", -1)

	err = json.Unmarshal([]byte(replyStr), &reply)
	if err != nil {
		return nil, err
	}
	return reply.Result, err
}
