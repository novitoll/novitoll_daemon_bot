package bot

import (
	"log"
	"strings"
	"time"
)

const (
	NEWCOMER_TIME_TO_RESPOND_BOT = 60
	NEWCOMER_RESPONSE = "pong"
)

var (
	botReplyMessage = "0!0 Ping, please write me 'pong' within 60 seconds, otherwise you will be kicked for a variety of reasons."
	chNewcomer = make(chan int)  // unbuffered chhanel to wait for the certain time for the newcomer's response
)

func (j *Job) actionOnNewMember() {
	log.Printf("[+] New member has been detected")

	// record a newcomer and wait for his reply on the channel,
	// otherwise kick that bastard and delete the record from this map
	NewComers[j.br.Message.From.Id] = time.Now()	

	botEgressReq := &BotEgressRequest{
		ChatId:					j.br.Message.Chat.Id,
		Text:					botReplyMessage,
		ParseMode:				ParseModeMarkdown,
		DisableWebPagePreview:	true,
		DisableNotification:	true,
		ReplyToMessageId:		j.br.Message.MessageId,
		ReplyMarkup:			&BotForceReply{ForceReply: true, Selective: true}}

	botEgressReq.EgressSendToTelegram(j.rh)
}

func JobNewChatMemberDetector(j *Job) (interface{}, error) {
	if j.br.Message.NewChatMember.Username == "" {
		return nil, nil
	}

	log.Println("[+] 11111111111111111")

	go j.actionOnNewMember()

	log.Println("[+] 22222222222222222")

	select {
	case dootId := <-chNewcomer:
		log.Printf("[+] Newcomer %d has been authenticated", dootId)
		delete(NewComers, j.br.Message.From.Id)
	case <-time.After(10 * time.Second):
		log.Printf("[!] Newcomer %s has been kicked", j.br.Message.NewChatMember.Username)
	}

	log.Println("[+] 33333333333")

	return nil, nil
}

// will check every message if its from a newcomer to whitelist the doot
func JobNewChatMemberWaiter(j *Job) (interface{}, error) {
	if _, ok := NewComers[j.br.Message.From.Id]; ok && strings.ToLower(j.br.Message.Text) == NEWCOMER_RESPONSE {
		chNewcomer <-j.br.Message.From.Id
	}
	return nil, nil
}