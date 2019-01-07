// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	cfg "github.com/novitoll/novitoll_daemon_bot/config"
	redisClient "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	admins          = []string{"@novitoll"}
	gopath, _       = os.LookupEnv("GOPATH")
	gopathPostfix   = "src/github.com/novitoll/novitoll_daemon_bot"
	testdataDirPath = "internal/vahter/bot/testdata"
)

func concatStringsWithSlash(s []string) string {
	return strings.Join(s[:], "/")
}

func filepathToStruct(t *testing.T, filepath string, pData interface{}) {
	var absolutePath string = concatStringsWithSlash([]string{gopath, gopathPostfix, filepath})
	fileJson, err := os.Open(absolutePath)
	defer fileJson.Close()

	if err != nil {
		panic(err)
		assert.Nil(t, err, "[-] Should be a valid existing file")
	}

	jsonBytes, _ := ioutil.ReadAll(fileJson)
	err = json.Unmarshal(jsonBytes, pData)
	assert.Nil(t, err, "[-] Should be valid features config JSON to decode")
}

func configureStructs(t *testing.T, reqBodyFilepath string) (*cfg.FeaturesConfig, *BotIngressRequest) {
	// FeaturesConfig init
	var features cfg.FeaturesConfig
	filepathToStruct(t, "config/features.json", &features)
	assert.Equal(t, features.Administration.Admins, admins, "[-] Should be equal FeaturesConfig struct features.Admins field")

	// BotIngressRequest init
	var ingressBody BotIngressRequest
	filepathToStruct(t, reqBodyFilepath, &ingressBody)

	return &features, &ingressBody
}

func TestDifferentIngressMessageStructs(t *testing.T) {
	files, err := ioutil.ReadDir("testdata")
	if err != nil {
		assert.Nil(t, err, "[-] Should be no error in loading testdata")
	}

	for _, f := range files {
		s := []string{testdataDirPath, f.Name()}
		configureStructs(t, concatStringsWithSlash(s))
	}
}

func TestURLDuplication(t *testing.T) {
	s := []string{testdataDirPath, "ingress_reqbody-url-1.json"}
	pFeatures, pBotRequest := configureStructs(t, concatStringsWithSlash(s))

	app := App{pFeatures, "en-us", logrus.New(), make(map[int]interface{})}
	pBotRequest.Process(&app)

	client := redisClient.GetRedisConnection()
	defer client.Close()
	client.FlushAll()
	expected := client.Get("https://test-123.com")

	assert.NotNilf(t, expected, "[-] Value from Redis should not be Nil found by key=extracted URL from message")
}

func TestTelegramResponseBodyStruct(t *testing.T) {
	s := []string{testdataDirPath, "ingress_responsebody-1.json"}
	_, pTelegramResponse := configureStructs(t, concatStringsWithSlash(s))
	assert.NotNilf(t, pTelegramResponse, "[-] Telegram response body should not be empty but valid")
}
