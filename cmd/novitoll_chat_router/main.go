// SPDX-License-Identifier: GPL-2.0
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/novitoll/novitoll_daemon_bot/internal/bot"
	"github.com/sirupsen/logrus"
)

type Router struct {
	Logger *logrus.Logger
}

const (
	CHAT_PORT = 8080
)

func (ro *Router) Route(w http.ResponseWriter, r *http.Request) {
	// parse the incoming HTTP request
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	buffer := new(bytes.Buffer)
	n, err := buffer.ReadFrom(r.Body)

	if n >= bot.MAX_ALLOWED_BUFFER_SIZE {
		ro.Logger.Warn("Exceeded max buffer size")
	}

	if err != nil {
		ro.Logger.Fatal(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var br bot.BotInReq
	err = json.Unmarshal([]byte(buffer.String()), &br)
	if err != nil {
		ro.Logger.Fatal(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// route to chat service per chat_id as TCP.
	// Service hostname is discovered as: http://bot_<chat_id>,
	// where chat_id is absolute value (chat_id can be signed int))
	var chat_id int64 = int64(math.Abs(float64(br.Message.Chat.Id)))

	service := fmt.Sprintf("http://bot_%d:%d/process", chat_id, CHAT_PORT)

	resp, err2 := http.Post(service, "application/json", buffer)
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
	http.ListenAndServe(":8080", nil)
}
