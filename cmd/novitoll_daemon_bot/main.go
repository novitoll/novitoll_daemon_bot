// SPDX-License-Identifier: GPL-2.0
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"

	cfg "github.com/novitoll/novitoll_daemon_bot/config"
	"github.com/novitoll/novitoll_daemon_bot/internal/vahter/bot"
	"github.com/sirupsen/logrus"
)

var (
	features cfg.FeaturesConfig
	lang     string = "en-us"
	logger          = logrus.New()
)

func init() {
	// 1. load features configuration
	fileJson, err := os.Open("config/features.json")
	if err != nil {
		logger.WithFields(logrus.Fields{"err": err}).Fatal("[-] Can not open features JSON")
	}
	defer fileJson.Close()

	featuresJsonBytes, _ := ioutil.ReadAll(fileJson)
	json.Unmarshal(featuresJsonBytes, &features)

	// 2. setup application i18n
	if l, ok := os.LookupEnv("APP_LANG"); ok {
		lang = l
	}
	if _, ok := features.NewcomerQuestionnare.I18n[lang]; !ok {
		msg := "Unknown language"
		logger.WithFields(logrus.Fields{"lang": lang}).Fatal(msg)
		panic(msg)
	}

	// 3. setup logger
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

func printReflectValues(s reflect.Value) {
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("-- %s %s = %v\n", typeOfT.Field(i).Name, f.Type(), f.Interface())

		if f.Kind().String() == "struct" {
			x1 := reflect.ValueOf(f.Interface())
			printReflectValues(x1)
			fmt.Printf("\n")
		}
	}
}

func main() {
	logger.Info("[+] Features config is loaded. Bot features:\n")

	featureFields := reflect.ValueOf(&features).Elem()
	printReflectValues(featureFields)

	handler := bot.App{
		Features:   &features,
		Lang:       lang,
		Logger:     logger,
		ChatAdmins: make(map[int][]string),
	}
	handler.RegisterHandlers()

	logger.Info("[+] Serving TCP 8080 port..")
	http.ListenAndServe(":8080", nil)
}
