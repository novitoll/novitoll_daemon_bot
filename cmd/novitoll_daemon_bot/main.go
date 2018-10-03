package main

import (
	"net/http"
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
	
	"github.com/novitoll/novitoll_daemon_bot/internal/vahter/router"
	cfg "github.com/novitoll/novitoll_daemon_bot/config"
)

func applyConfig() *cfg.FeaturesConfig {
	var features cfg.FeaturesConfig	
	
	if file, err :+ os.Open("config/features.json"); err != nil {
		panic(err)
	}
	defer file.close()

	featuresJsonBytes, _ := iotuil.ReadAll(file)

	json.Unmarshal(featuresJsonBytes, &features)

	return &features
}

func main() {
	features := applyConfig()
	handler := router.RouteHandler{&features}
	handler.RegisterHandlers()

	http.ListenAndServe(":8080", nil)
}
