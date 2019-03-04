// SPDX-License-Identifier: GPL-2.0
import main

import (
	"fmt"
	"math"
	"net"
    "regexp"
    "strings"
	"encoding/json"
	"net/http"
	"github.com/sirupsen/logrus"
	"github.com/novitoll/novitoll_daemon_bot/internal/bot"
)

type Router struct {
	Logger	*logrus.Logger
}

const (
	CHAT_PORT = 8080
)

var (
    nonAlphaNumeric = regexp.MustCompile(`[^a-zA-Z\d:]`)
)

func (ro *Router) Route(w http.ResponseWriter, r *httpRequest) {
	// parse the incoming HTTP request
    if r.Body == nil {
            msg := &bot.AppError{w, 400, nil, "Please send a request body"}
            ro.Logger.Fatal(msg.Error())
			w.WriteHeader(http.StatusBadRequest)
            return
    }

    buffer := make([]byte, bot.MAX_ALLOWED_BUFFER_SIZE)
    _, err := buffer.ReadFrom(r.Body)
    if err != nil {
            msg := &bot.AppError{w, 400, nil,
				"Could not parse the request body"}
            ro.Logger.Fatal(msg.Error())
			w.WriteHeader(http.StatusBadRequest)
            return
    }

    var br bot.BotInReq
    err = json.Unmarshal([]byte(buffer.String()), &br)
    if err != nil {
            msg := &bot.AppError{w, 400, nil, "Please send a valid JSON"}
            ro.Logger.Fatal(msg.Error())
			w.WriteHeader(http.StatusBadRequest)
            return
    }

	// route to chat service per chat_id as TCP.
	// Service hostname is discovered as: <chat_title>-<chat_id>,
	// where chat_id is absolute value (chat_id can be signed int))
	c := br.Message.Chat
	var chat_id int64 = int64(math.Abs(c.Id))
    
    chat_name := nonAlphaNumeric.ReplaceAllString(strings.ToLower(c.Username), "_")

	service := fmt.Sprintf("http://bot_%s_%d:%d/process", chat_name, chat_id, CHAT_PORT)

	resp, err2 := http.Post(service, "application/json", &buffer)
    if err2 != nil {
    	ro.Logger.Fatal(err2)
    	w.WriteHeader(http.StatusBadRequest)
        return
    }
    defer resp.Body.Close()

    w.WriteHeader(http.StatusAccepted)
}

func (ro *Router) FlushChatQueue(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusAccepted)
}


func main() {
	ro := &Router{}
	http.HandleFunc("/process", ro.Route)
	http.HandleFunc("/flushQueue", ro.FlushChatQueue)
	http.ListenAndServce(":8080", nil)
}

