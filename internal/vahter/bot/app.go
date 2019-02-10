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
	// This global variable is referenced in cronjobs.go
	ChatIds = make(map[int]time.Time)
)

type App struct {
	Features   *cfg.FeaturesCfg
	Lang       string
	Logger     *logrus.Logger
	ChatAdmins map[int][]string
}

func (app *App) RegisterHandlers() {
	http.HandleFunc("/process", app.ProcessHandler)
	http.HandleFunc("/flushQueue", app.FlushQueueHandler)

	app.Logger.Info("[+] Handlers for HTTP end-points " + 
		"are registered")
}

// Receives HTTP requests on /process end-point
// from Telegram and after parsing request body raw bytes,
// "BotInReq" struct is created which contains 
// sufficient info about Telegram Chat, User, Message.
// After struct's creation, Process() goroutine is scheduled and
// the HTTP request handler is completed with 200/OK response.
func (app *App) ProcessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		msg := &AppError{w, 400, nil, "Please send a request body"}
		app.Logger.Fatal(msg.Error())
		return
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		msg := &AppError{w, 400, nil, "Could not parse " +
			"the request body"}
		app.Logger.Fatal(msg.Error())
		return
	}
	// This should be useful to analyze STDOUT logs
	// instead of storing them in RDBMS etc.
	app.Logger.Info(buf.String())

	var br BotInReq
	err = json.Unmarshal([]byte(buf.String()), &br)
	if err != nil {
		msg := &AppError{w, 400, nil, "Please send a valid JSON"}
		app.Logger.Fatal(msg.Error())
		return
	}

	go br.Process(app)

	// we cant run crons unless we know chat ID
	if _, ok := ChatIds[br.Message.Chat.Id]; !ok {
		ChatIds[br.Message.Chat.Id] = time.Now()
		go br.CronSchedule(app)

		app.Logger.Info(fmt.Sprintf("[+] Cron jobs for %d chat " +
			"are scheduled", br.Message.Chat.Id))
	}

	w.WriteHeader(http.StatusAccepted)
}

func (app *App) FlushQueueHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}