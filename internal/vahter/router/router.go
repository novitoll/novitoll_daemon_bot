package router

import 	(
	"log"
	"net/http"
	"encoding/json"

	"github.com/novitoll/novitoll_daemon_bot/internal/vahter/bot"
	cfg "github.com/novitoll/novitoll_daemon_bot/config"
)

type RouteHandler struct {
	features *cfg.FeaturesConfig
}

func (rh *RouteHandler) RegisterHandlers() {
	http.HandleFunc("/check", handlerWrapper(CheckHandler))
	// http.HandleFunc("/status", handlerWrapper(StatusHandler))
}

func (rh *RouteHandler) CheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		&RouteError{w, 400, nil, "Please send a request body"}
		return
	}

	var br bot.BotRequest
	err := json.NewDecoder(r.Body).Decode(&br)
	if err != nil {
		&RouteError{w, 400, nil, "Please send a valid JSON"}
		return
	}

	go br.Process(&rh)

    w.WriteHeader(http.StatusOK)
}

// func (h *RouteHandler) StatusHandler(w http.ResponseWriter, r *http.Request) {

// }
