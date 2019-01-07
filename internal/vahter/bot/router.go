// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"bytes"
	"encoding/json"
	"net/http"

	cfg "github.com/novitoll/novitoll_daemon_bot/config"
	"github.com/sirupsen/logrus"
)

type App struct {
	Features *cfg.FeaturesConfig
	Lang     string
	Logger   *logrus.Logger
}

func (app *App) RegisterHandlers() {
	http.HandleFunc("/process", app.ProcessMessageHandlerFunc)
	http.HandleFunc("/flushQueue", app.FlushQueueHandlerFunc)

	app.Logger.Info("[+] Handlers for HTTP end-points are registered")
}

func (app *App) RegisterCronJobs() {
	go func() {
		app.StartCronJobs()
	}()

	app.Logger.Info("[+] Cronjobs are scheduled")
}

func (app *App) ProcessMessageHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		msg := &RouteError{w, 400, nil, "Please send a request body"}
		app.Logger.Fatal(msg.Error())
		return
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		msg := &RouteError{w, 400, nil, "Could not parse the request body"}
		app.Logger.Fatal(msg.Error())
		return
	}
	app.Logger.Info(buf.String()) // TODO: remove

	var br BotIngressRequest
	err = json.Unmarshal([]byte(buf.String()), &br)
	if err != nil {
		msg := &RouteError{w, 400, nil, "Please send a valid JSON"}
		app.Logger.Fatal(msg.Error())
		return
	}

	go br.Process(app)
	w.WriteHeader(http.StatusAccepted)
}

func (app *App) FlushQueueHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}
