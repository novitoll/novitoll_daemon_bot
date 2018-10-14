package bot

import (
	"log"
	"strings"
	"time"
)

const (
	NEWCOMER_TIME_TO_RESPOND_BOT = 60
	NEWCOMER_RESPONSE = "pong"
	NEWCOMER_KICK_TIME = 60
)

var (
	botReplyMessage = "Ping, please write me 'pong' within 60 seconds, otherwise you will be kicked for a variety of reasons. #novitollnm"
	chNewcomer = make(chan int)  // unbuffered chhanel to wait for the certain time for the newcomer's response
)

func JobNewChatMemberDetector(j *Job) (interface{}, error) {
	if j.br.Message.NewChatMember.Username == "" {
		return nil, nil
	}

	go j.actionSendMessage(botReplyMessage, false)

	select {
	case dootId := <-chNewcomer:		
		delete(NewComers, dootId)
		log.Printf("[+] Newcomer %d has been authenticated", dootId)

		if j.rh.Features.NewcomerQuestionnare.ActionNotify {
			j.actionSendMessage("Thanks. You are whitelisted #novitollwl", true)
		}
		
	case <-time.After(NEWCOMER_TIME_TO_RESPOND_BOT * time.Second):
		if kicked := j.actionKickChatMember(); kicked {
			delete(NewComers, j.br.Message.NewChatMember.Id)
			log.Printf("[!] Newcomer %s has been kicked", j.br.Message.NewChatMember.Username)
		}
	}

	return nil, nil
}

func JobNewChatMemberWaiter(j *Job) (interface{}, error) {
	// will check every message if its from a newcomer to whitelist the doot, writing to the global unbuffered channel
	if _, ok := NewComers[j.br.Message.From.Id]; ok && strings.ToLower(j.br.Message.Text) == NEWCOMER_RESPONSE {
		chNewcomer <-j.br.Message.From.Id
	}
	return nil, nil
}

func (j *Job) actionSendMessage(text string, isAuth bool) {
	if !isAuth {
		// record a newcomer and wait for his reply on the channel,
		// otherwise kick that bastard and delete the record from this map
		log.Printf("[+] New member has been detected")
		NewComers[j.br.Message.NewChatMember.Id] = time.Now()
	}

	botEgressReq := &BotEgressRequest{
		ChatId:					j.br.Message.Chat.Id,
		Text:					text,
		ParseMode:				ParseModeMarkdown,
		DisableWebPagePreview:	true,
		DisableNotification:	true,
		ReplyToMessageId:		j.br.Message.MessageId,
		ReplyMarkup:			&BotForceReply{ForceReply: true, Selective: true}}

	botEgressReq.EgressSendToTelegram(j.rh)
}

func (j *Job) actionKickChatMember() bool {
	log.Printf("[+] Kicking a newcomer")

	t := time.Unix(j.br.Message.Date, 0)

	botEgressReq := &BotEgressKickChatMember{
		ChatId: j.br.Message.Chat.Id,
		UserId: j.br.Message.NewChatMember.Id,
		UntilDate: t.Add(NEWCOMER_KICK_TIME * time.Second).Unix(),
	}

	return botEgressReq.EgressKickChatMember(j.rh)
}