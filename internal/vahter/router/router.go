package router

import (
	// "log"
	"net/http"
	"encoding/json"

	"github.com/novitoll/novitoll_daemon_bot/internal/vahter/bot"
)

type RouteHandler struct {}

func (h *RouteHandler) FaqHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	var br bot.BotRequest
	err := json.NewDecoder(r.Body).Decode(&br)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	br.Process()

    // w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    // w.Write("cool")
}
