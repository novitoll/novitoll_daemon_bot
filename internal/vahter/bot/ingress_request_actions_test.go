package bot

import (
	"os"
	"time"
	"strings"
	"testing"
	"io/ioutil"
	"encoding/json"
	"github.com/stretchr/testify/assert"

	cfg "github.com/novitoll/novitoll_daemon_bot/config"
	redisClient "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
)

var (
	admins = []string{"novitoll"}
	gopath, _ = os.LookupEnv("GOPATH")
	gopathPostfix = "src/github.com/novitoll/novitoll_daemon_bot"
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
	assert.Equal(t, features.NotificationTarget.Admins, admins, "[-] Should be equal FeaturesConfig struct features.Admins field")

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
		s := []string{"internal/vahter/bot/testdata", f.Name()}
		configureStructs(t, concatStringsWithSlash(s))
	}	
}

func TestURLDuplication(t *testing.T) {
	reqBodyFilepath := "internal/vahter/bot/testdata/ingress_reqbody-url-1.json"
	pFeatures, pBotRequest := configureStructs(t, reqBodyFilepath)

	app := App{pFeatures, "en-us"}
	pBotRequest.Process(&app)

	client := redisClient.GetRedisConnection()
	defer client.Close()
	client.FlushAll()
	expected := client.Get("https://test-123.com")

	assert.NotNilf(t, expected, "[-] Value from Redis should not be Nil found by key=extracted URL from message")
}

func TestNewComer(t *testing.T) {
	reqBodyFilepath := "internal/vahter/bot/testdata/ingress_reqbody-newchatmember-1.json"
	_, pBotRequest := configureStructs(t, reqBodyFilepath)

	timer := time.NewTimer(11 * time.Second)
	go func() {
		<-timer.C
		assert.Contains(t, NewComers, pBotRequest.Message.From.Id, "[-] Should be NewComers map")
	}()
}