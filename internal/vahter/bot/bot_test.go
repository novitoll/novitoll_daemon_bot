package bot

import (
	"testing"
	"encoding/json"

	"github.com/stretchr/testify/assert"
	redisClient "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
	cfg "github.com/novitoll/novitoll_daemon_bot/config"
)

func configureStructs(t *testing.T) (*cfg.FeaturesConfig, *BotRequest) {
	var featuresConfig, botRequest string

	featuresConfig = `
		{"admins": ["novitoll"],
			"urlDuplication": {
			"enabled": true,
			"actionKick": false,
			"actionBan": false,
			"actionAdminNotify": true
			},
			"newcomerQuestionnare": {
			"enabled": true,
			"actionKick": false,
			"actionBan": false,
			"actionAdminNotify": true
			},
			"adDetection": {
			"enabled": true,
			"actionKick": false,
			"actionBan": false,
			"actionAdminNotify": true
		}}`

	var features cfg.FeaturesConfig
	err := json.Unmarshal([]byte(featuresConfig), &features)
	assert.Nil(t, err, "[-] Should be valid features config JSON to decode")

	admins := []string{"novitoll"}
	assert.Equal(t, features.Admins, admins, "[-] Should be equal FeaturesConfig struct features.Admins field")

	botRequest = `
		{"message": {
		"from": {
		  "username": "novitoll",
		  "first_name": "novitoll",
		  "is_bot": false,
		  "id": 345019684,
		  "language_code": "en-US"
		},
		"text": "message https://test-123.com",
		"entities": [
		  {
		    "length": 101,
		    "type": "url",
		    "offset": 0
		  }
		],
		"chat": {
		  "username": "novitoll",
		  "first_name": "novitoll",
		  "type": "private",
		  "id": 345019684
		},
		"date": 1537020424,
		"message_id": 28
		},
		"update_id": 776799951
		}`

	var br BotRequest
	err2 := json.Unmarshal([]byte(botRequest), &br)
	assert.Nil(t, err2, "[-] Should be valid BotRequest JSON to decode")

	assert.Equal(t, br.Message.From.Username, "novitoll", "[-] Should be equal decoded BotRequest struct Message.From.Username field")

	return &features, &br
}

func TestURLDuplication(t *testing.T) {
	pFeatures, pBotRequest := configureStructs(t)

	rh := RouteHandler{pFeatures}
	pBotRequest.Process(&rh)

	client := redisClient.GetRedisConnection()
	expected := client.Get("https://test-123.com")

	assert.NotNilf(t, expected, "[-] Value from Redis should not be Nil found by key=extracted URL from message")
}
