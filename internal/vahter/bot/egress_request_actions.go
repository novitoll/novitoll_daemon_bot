package bot

import (
	"fmt"
	"bytes"
	"time"
	"net"
	"net/http"
	"encoding/json"
)

func (reqBody *BotEgressRequest) EgressSendToTelegram(rh *RouteHandler) {
	jsonValue, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(TELEGRAM_URL, rh.Features.NotificationTarget.Token)
	req, err := http.NewRequest(POST, url, bytes.NewBuffer(jsonValue))

	req.Header.Set("Content-Type", "application/json")

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: netTransport,	
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}