// SPDX-License-Identifier: GPL-2.0
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"

	cfg "github.com/novitoll/novitoll_daemon_bot/config"
	"github.com/novitoll/novitoll_daemon_bot/internal/bot"
	"github.com/novitoll/novitoll_daemon_bot/pkg/utils"
	"github.com/sirupsen/logrus"
)

// Global variables of "main" pkg
var (
	features	 cfg.FeaturesCfg
	lang		 = "en-us"
	logger           = logrus.New()
	requestID        = 0
)

// Initializes FeaturesCfg and do following configuration:
// 1. load features configuration
// 2. setup application i18n
// 3. setup logger
func init() {
	// 1.
	fileJson, err := os.Open("config/features.json")
	if err != nil {
		logger.WithFields(logrus.Fields{"err": err}).
			Fatal("[-] Can not open features JSON")
	}
	defer fileJson.Close()

	featuresJsonBytes, _ := ioutil.ReadAll(fileJson)
	json.Unmarshal(featuresJsonBytes, &features)

	// 2.
	if l, ok := os.LookupEnv("APP_LANG"); ok {
		lang = l
	}

	if _, ok := features.NewcomerQuestionnare.I18n[lang]; !ok {
		msg := "Unknown language"
		logger.WithFields(logrus.Fields{"lang": lang}).Fatal(msg)
		panic(msg)
	}

	// 3.
	logger.SetOutput(os.Stdout)

	switch features.Administration.LogLevel {
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
		break
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
}

func nextRequestID() int {
	requestID++
	return requestID
}

// Starts HTTP server based on "net/http" pkg on TCP/8080 constant port.
// Prints to STDOUT configuration from features.json.
// Creates the only App{} struct which will be used along the way, and
// while struct fields remain constant, Apps->ChatAdmins field will be
// changed per each cron job (see cronjobs.go)
func main() {
	logger.Info("[+] Features config is loaded. Bot features:\n")

	featureFields := reflect.ValueOf(&features).Elem()
	utils.PrintReflectValues(featureFields)

	app := bot.App{
		Features:   &features,
		Lang:       lang,
		Logger:     logger,
		ChatAdmins: make(map[int][]string),
	}
	app.RegisterHandlers()

	logger.Info("[+] Serving TCP 8080 port..")
	http.ListenAndServe(":8080", nil)
}
