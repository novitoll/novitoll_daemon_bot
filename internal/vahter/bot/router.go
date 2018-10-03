package bot

import 	(
	"log"
	"net/http"
	"encoding/json"

	cfg "github.com/novitoll/novitoll_daemon_bot/config"
)

type RouteHandler struct {
	Features *cfg.FeaturesConfig
}

func (rh *RouteHandler) RegisterHandlers() {
	http.HandleFunc("/check", rh.CheckHandlerFunc)

	log.Printf("[+] Handlers for HTTP end-points are registered")
}

/*
	Handlers
*/
func (rh *RouteHandler) CheckHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		msg := &RouteError{w, 400, nil, "Please send a request body"}
		log.Fatalf(msg.Error())
		return
	}

	var br BotRequest
	err := json.NewDecoder(r.Body).Decode(&br)
	if err != nil {
		msg := &RouteError{w, 400, nil, "Please send a valid JSON"}
		log.Fatalf(msg.Error())
		return
	}

	go br.Process(rh)

    w.WriteHeader(http.StatusOK)
}
