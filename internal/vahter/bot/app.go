// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cfg "github.com/novitoll/novitoll_daemon_bot/config"
	"github.com/sirupsen/logrus"
)

var (
	ChatIds = make(map[int]time.Time)
)

type App struct {
	Features *cfg.FeaturesConfig
	Lang     string
	Logger   *logrus.Logger
	ChatAdmins map[int][]string
}

func (app *App) RegisterHandlers() {
	http.HandleFunc("/process", app.ProcessMessageHandlerFunc)
	http.HandleFunc("/flushQueue", app.FlushQueueHandlerFunc)

	app.Logger.Info("[+] Handlers for HTTP end-points are registered")
}

func (app *App) ProcessMessageHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		msg := &AppError{w, 400, nil, "Please send a request body"}
		app.Logger.Fatal(msg.Error())
		return
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		msg := &AppError{w, 400, nil, "Could not parse the request body"}
		app.Logger.Fatal(msg.Error())
		return
	}
	app.Logger.Info(buf.String()) // TODO: remove

	var br BotIngressRequest
	err = json.Unmarshal([]byte(buf.String()), &br)
	if err != nil {
		msg := &AppError{w, 400, nil, "Please send a valid JSON"}
		app.Logger.Fatal(msg.Error())
		return
	}

	go br.Process(app)

	// we cant run crons unless we know chat ID
	if _, ok := ChatIds[br.Message.Chat.Id]; !ok {
		app.Logger.Info(fmt.Sprintf("[+] Cron jobs for %d chat are scheduled", br.Message.Chat.Id))
		ChatIds[br.Message.Chat.Id] = time.Now()
		go br.CronJobsStartForChat(app)
	}

	w.WriteHeader(http.StatusAccepted)
}

func (app *App) FlushQueueHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}
