package bot

import (
	"log"
	"strings"
	"time"
)

var (
	botReplyMessage = "Ping, please write me *pong* within 60 seconds, otherwise you will be kicked for a variety of reasons. #novitollnm"
	chNewcomer = make(chan int)  // unbuffered chhanel to wait for the certain time for the newcomer's response
)

func JobNewChatMemberDetector(j *Job) (interface{}, error) {
	// for short code reference
	newComer := j.ingressBody.Message.NewChatMember
	newComerConfig := j.app.Features.NewcomerQuestionnare

	if newComer.Id == 0 {
		return nil, nil
	}

	go j.actionSendMessage(botReplyMessage, false)

	// blocks the current Job goroutine until either of these 2 channels receive the value
	select {
	case dootId := <-chNewcomer:
		delete(NewComers, dootId)
		log.Printf("[+] Newcomer %d has been authenticated", dootId)

		if newComerConfig.ActionNotify {
			j.actionSendMessage("Thanks. You are whitelisted #novitollwl", true)
		}		
	case <-time.After(time.Duration(newComerConfig.AuthTimeout) * time.Second):
		if kicked := j.actionKickChatMember(); kicked {
			delete(NewComers, newComer.Id)
			log.Printf("[!] Newcomer %s has been kicked", newComer.Username)
		}
	}

	return nil, nil
}

func JobNewChatMemberWaiter(j *Job) (interface{}, error) {
	// will check every message if its from a newcomer to whitelist the doot, writing to the global unbuffered channel
	if _, ok := NewComers[j.ingressBody.Message.From.Id]; ok && strings.ToLower(j.ingressBody.Message.Text) == j.app.Features.NewcomerQuestionnare.AuthMessage {
		chNewcomer <-j.ingressBody.Message.From.Id
	}
	return nil, nil
}

func (j *Job) actionSendMessage(text string, isAuth bool) {
	if !isAuth {
		// record a newcomer and wait for his reply on the channel,
		// otherwise kick that bastard and delete the record from this map
		log.Printf("[+] New member has been detected")
		NewComers[j.ingressBody.Message.NewChatMember.Id] = time.Now()
	}

	botEgressReq := &BotEgressSendMessage{
		ChatId:					j.ingressBody.Message.Chat.Id,
		Text:					text,
		ParseMode:				ParseModeMarkdown,
		DisableWebPagePreview:	true,
		DisableNotification:	true,
		ReplyToMessageId:		j.ingressBody.Message.MessageId,
		ReplyMarkup:			&BotForceReply{ForceReply: true, Selective: true}}

	botEgressReq.EgressSendToTelegram(j.app)
}

func (j *Job) actionKickChatMember() bool {
	log.Printf("[+] Kicking a newcomer")

	t := time.Unix(j.ingressBody.Message.Date, 0)

	botEgressReq := &BotEgressKickChatMember{
		ChatId: j.ingressBody.Message.Chat.Id,
		UserId: j.ingressBody.Message.NewChatMember.Id,
		UntilDate: t.Add(time.Duration(j.app.Features.NewcomerQuestionnare.KickBanTimeout) * time.Second).Unix(),
	}

	return botEgressReq.EgressKickChatMember(j.app)
}