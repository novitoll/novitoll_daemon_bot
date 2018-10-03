package main

import (
	"os"
	"log"
	"net/http"
	"io/ioutil"
	"encoding/json"
	
	"github.com/novitoll/novitoll_daemon_bot/internal/vahter/bot"
	cfg "github.com/novitoll/novitoll_daemon_bot/config"
)

func getFeaturesConfig() *cfg.FeaturesConfig {
	var features cfg.FeaturesConfig		
	fileJson, err := os.Open("config/features.json")
	if err != nil {
		panic(err)
	}
	defer fileJson.Close()

	featuresJsonBytes, _ := ioutil.ReadAll(fileJson)
	json.Unmarshal(featuresJsonBytes, &features)
	return &features
}

func main() {
	features := getFeaturesConfig()
	log.Printf("[+] Features config is loaded. Enabled features: - %t", features.UrlDuplication.Enabled)

	handler := bot.RouteHandler{features}
	handler.RegisterHandlers()

	log.Printf("[+] Serving TCP 8080 port..")
	http.ListenAndServe(":8080", nil)	
}
