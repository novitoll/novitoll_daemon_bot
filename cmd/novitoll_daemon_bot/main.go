package main

import (
	"os"
	"log"
	"fmt"
	"reflect"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/novitoll/novitoll_daemon_bot/internal/vahter/bot"
	cfg "github.com/novitoll/novitoll_daemon_bot/config"
)

var (
	features cfg.FeaturesConfig
)

func init() {
	fileJson, err := os.Open("config/features.json")
	if err != nil {
		log.Fatalln("[-] Can not open features JSON\n", err)
	}
	defer fileJson.Close()

	featuresJsonBytes, _ := ioutil.ReadAll(fileJson)
	json.Unmarshal(featuresJsonBytes, &features)
}

func printReflectValues(v reflect.Value) {
	s := v
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
	log.Printf("[+] Features config is loaded. Bot features:\n")

	featureFields := reflect.ValueOf(&features).Elem()
	printReflectValues(featureFields)

	handler := bot.RouteHandler{&features}
	handler.RegisterHandlers()

	log.Printf("[+] Serving TCP 8080 port..")
	http.ListenAndServe(":8080", nil)
}
