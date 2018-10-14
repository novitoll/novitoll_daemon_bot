package bot

import (
	"testing"
	"time"
	"encoding/json"

	"github.com/stretchr/testify/assert"
	redisClient "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
	cfg "github.com/novitoll/novitoll_daemon_bot/config"
)

var (
	botRequestMessageMigration = `
	{"update_id":53205698, "message":{"message_id":108,"from":{"id":345019684,"is_bot":false,"first_name":"novitoll","username":"novitoll","language_code":"en-US"},"chat":{"id":-253761934,"title":"test_novitoll_daemon_bot","type":"group","all_members_are_administrators":true},"date":1539515199,"migrate_to_chat_id":-1001276148791}}
	`
	
	botRequestNewComer = `
	{"update_id":53205697, "message":{"message_id":107,"from":{"id":345019684,"is_bot":false,"first_name":"novitoll","username":"novitoll","language_code":"en-US"},"chat":{"id":-253761934,"title":"test_novitoll_daemon_bot","type":"group","all_members_are_administrators":true},"date":1539515176,"new_chat_participant":{"id":553713145,"is_bot":true,"first_name":"novitoll_daemon_bot","username":"novitoll_daemon_bot"},"new_chat_member":{"id":553713145,"is_bot":true,"first_name":"novitoll_daemon_bot","username":"novitoll_daemon_bot"},"new_chat_members":[{"id":553713145,"is_bot":true,"first_name":"novitoll_daemon_bot","username":"novitoll_daemon_bot"}]}}
	`

	botRequestPlainText = `
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

	botRequestMemberLeftChat = `
	{"update_id":53205716, "message":{"message_id":25,"from":{"id":345019684,"is_bot":false,"first_name":"novitoll","username":"novitoll","language_code":"en-US"},"chat":{"id":-1001276148791,"title":"test_novitoll_daemon_bot","type":"supergroup"},"date":1539530333,"left_chat_participant":{"id":574825208,"is_bot":false,"first_name":"N2","username":"Novitoll_2"},"left_chat_member":{"id":574825208,"is_bot":false,"first_name":"N2","username":"Novitoll_2"}}}
	`
)

func configureStructs(t *testing.T) (*cfg.FeaturesConfig, *BotIngressRequest) {
	var featuresConfig string

	featuresConfig = `
	{
	  "notificationTarget": {
	    "admins": ["novitoll"]
	  },
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
	  }
	}`

	var features cfg.FeaturesConfig
	err := json.Unmarshal([]byte(featuresConfig), &features)
	assert.Nil(t, err, "[-] Should be valid features config JSON to decode")

	admins := []string{"novitoll"}
	assert.Equal(t, features.NotificationTarget.Admins, admins, "[-] Should be equal FeaturesConfig struct features.Admins field")

	var br BotIngressRequest
	err2 := json.Unmarshal([]byte(botRequestPlainText), &br)
	assert.Nil(t, err2, "[-] Should be valid BotIngressRequest JSON to decode")

	assert.Equal(t, br.Message.From.Username, "novitoll", "[-] Should be equal decoded BotIngressRequest struct Message.From.Username field")

	return &features, &br
}

func TestURLDuplication(t *testing.T) {
	pFeatures, pBotRequest := configureStructs(t)

	rh := RouteHandler{pFeatures}
	pBotRequest.Process(&rh)

	client := redisClient.GetRedisConnection()
	defer client.Close()
	client.FlushAll()
	expected := client.Get("https://test-123.com")

	assert.NotNilf(t, expected, "[-] Value from Redis should not be Nil found by key=extracted URL from message")
}

func TestDifferentIngressMessageStructs(t *testing.T) {
	var br BotIngressRequest
	err := json.Unmarshal([]byte(botRequestMessageMigration), &br)
	assert.Nil(t, err, "[-] Should be valid BotIngressRequest JSON to decode")

	var br2 BotIngressRequest
	err2 := json.Unmarshal([]byte(botRequestNewComer), &br2)
	assert.Nil(t, err2, "[-] Should be valid BotIngressRequest JSON to decode")
}

func TestNewComer(t *testing.T) {
	var br BotIngressRequest
	err := json.Unmarshal([]byte(botRequestNewComer), &br)
	assert.Nil(t, err, "[-] Should be valid BotIngressRequest JSON to decode")

	timer := time.NewTimer(11 * time.Second)
    go func() {
        <-timer.C
    	assert.Contains(t, NewComers, br.Message.From.Id, "[-] Should be NewComers map")
    }()
}