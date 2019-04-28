// SPDX-License-Identifier: GPL-2.0
package bot_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	cfg "github.com/novitoll/novitoll_daemon_bot/config"
	. "github.com/novitoll/novitoll_daemon_bot/internal/bot"
	redisClient "github.com/novitoll/novitoll_daemon_bot/pkg/redis_client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	gopath, _       = os.LookupEnv("GOPATH")
	gopathPostfix   = "src/github.com/novitoll/novitoll_daemon_bot"
	testdataDirPath = "internal/bot/testdata"
)

// ----------------------------------
// Helpers
// ----------------------------------

func ConcatStringsWithSlash(s []string) string {
	return strings.Join(s[:], "/")
}

func FilepathToStruct(t *testing.T, filepath string, pData interface{}) {
	var absolutePath string = ConcatStringsWithSlash([]string{gopath, gopathPostfix, filepath})
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

func ConfigureStructs(t *testing.T, reqBodyFilepath string) (*cfg.FeaturesCfg, *BotInReq) {
	// FeaturesCfg init
	var features cfg.FeaturesCfg
	FilepathToStruct(t, "config/features.json", &features)
	assert.Equal(t, features.Administration.LogLevel, "info",
	 "[-] Should be equal FeaturesCfg struct features.LogLevel field")

	// BotInReq init
	var req BotInReq
	FilepathToStruct(t, reqBodyFilepath, &req)

	return &features, &req
}

func BuildApp(t *testing.T) (*App, *BotInReq) {
	s := []string{testdataDirPath, "ingress_reqbody-url-1.json"}
	pFeatures, pBotRequest := ConfigureStructs(t, ConcatStringsWithSlash(s))

	mux := http.NewServeMux()

	app := &App{
		Features:   pFeatures,
		Lang:       "en-us",
		Logger:     logrus.New(),
		ChatAdmins: make(map[int][]string),
		Mux:        mux,
	}
	return app, pBotRequest
}

// ----------------------------------
// unit-tests: misc
// ----------------------------------

func TestDifferentIngressMessageStructs(t *testing.T) {
	files, err := ioutil.ReadDir("testdata")
	if err != nil {
		assert.Nil(t, err, "[-] Should be no error in loading testdata")
	}

	for _, f := range files {
		s := []string{testdataDirPath, f.Name()}
		ConfigureStructs(t, ConcatStringsWithSlash(s))
	}
}

func TestTelegramResponseBodyStruct(t *testing.T) {
	s := []string{testdataDirPath, "ingress_responsebody-1.json"}
	_, pTelegramResponse := ConfigureStructs(t, ConcatStringsWithSlash(s))
	assert.NotNilf(t, pTelegramResponse, "[-] Telegram response body should not be empty but valid")
}

// ----------------------------------
// unit-tests: handlers
// ----------------------------------

func TestHandlers(t *testing.T) {
	// ProcessHandler
	s := []string{testdataDirPath, "ingress_reqbody-1.json"}
	_, pBotRequest := ConfigureStructs(t, ConcatStringsWithSlash(s))

	jsonBytes, err := json.Marshal(pBotRequest)
	if err != nil {
		t.Fatal(err)
	}
	j := bytes.NewReader(jsonBytes)

	req, err2 := http.NewRequest("POST", "/process", j)
	if err2 != nil {
		t.Fatal(err2)
	}

	app, _ := BuildApp(t)
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.ProcessHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusAccepted)
	}

	// FlushQueueHandler
	req, err2 = http.NewRequest("POST", "/flushQueue", j)
	if err2 != nil {
		t.Fatal(err2)
	}
	handler = http.HandlerFunc(app.FlushQueueHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusAccepted)
}

/*
	Jobs' tests
	[-] JobNewChatMemberDetector,
	[-] JobNewChatMemberAuth,
	[..] JobUrlDuplicationDetector,
	[-] JobMsgStats,
	[-] JobAdDetector,
	[-] JobSentimentDetector,
	[-] JobLeftParticipantDetector,
*/

func TestJobURLDuplicationDetector(t *testing.T) {
	app, pBotRequest := BuildApp(t)

	pBotRequest.Process(app)

	client := redisClient.GetRedisConnection()
	defer client.Close()
	client.FlushAll()
	expected := client.Get("https://test-123.com")

	assert.NotNilf(t, expected, "[-] Value from Redis should not be Nil found by key=extracted URL from message")
}
