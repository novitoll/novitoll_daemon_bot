package bot

import (
	"log"
	"bytes"
	"net/http"
	"encoding/json"

	cfg "github.com/novitoll/novitoll_daemon_bot/config"
)

type RouteHandler struct {
	Features *cfg.FeaturesConfig
}

func (rh *RouteHandler) RegisterHandlers() {
	http.HandleFunc("/check", rh.CheckMessageHandlerFunc)

	log.Printf("[+] Handlers for HTTP end-points are registered")
}

func (rh *RouteHandler) CheckMessageHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		msg := &RouteError{w, 400, nil, "Please send a request body"}
		log.Fatalln(msg.Error())
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	log.Println(buf.String()) // TODO: remove

	var br BotIngressRequest
	err := json.Unmarshal([]byte(buf.String()), &br)
	if err != nil {
		msg := &RouteError{w, 400, nil, "Please send a valid JSON"}
		log.Fatalln(msg.Error())
		return
	}

	go br.Process(rh)
    w.WriteHeader(http.StatusAccepted)
}