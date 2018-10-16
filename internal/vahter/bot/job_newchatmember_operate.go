package bot

import (
	"log"
	"fmt"
	"time"
)

var (
	chNewcomer = make(chan int)  // unbuffered chhanel to wait for the certain time for the newcomer's response
)

// TODO: this method is too complex, make it more lightweight

func JobNewChatMemberDetector(j *Job) (bool, error) {
	// for short code reference
	newComer := j.ingressBody.Message.NewChatMember
	newComerConfig := j.app.Features.NewcomerQuestionnare
	botReplyMsg := newComerConfig.I18n[j.app.Lang]
	btnMsg := botReplyMsg.AuthMessage

	if !newComerConfig.Enabled || newComer.Id == 0 || newComer.Username == "@novitoll_daemon_bot" {
		return false, nil
	}

	keyBtns := [][]KeyboardButton{
		[]KeyboardButton{
			KeyboardButton{btnMsg},
		},
	}

	welcomeMsg := fmt.Sprintf(botReplyMsg.WelcomeMessage, newComerConfig.AuthTimeout, newComerConfig.KickBanTimeout)

	// record a newcomer and wait for his reply on the channel,
	// otherwise kick that bastard and delete the record from this map
	log.Printf("[+] New member %d(@%s) has been detected", newComer.Id, newComer.Username)
	NewComers[j.ingressBody.Message.NewChatMember.Id] = time.Now()

	go j.actionSendMessage(welcomeMsg, &ReplyKeyboardMarkup{keyBtns, true, true})

	// blocks the current Job goroutine until either of these 2 channels receive the value
	select {
	case dootId := <-chNewcomer:
		delete(NewComers, dootId)
		log.Printf("[+] Newcomer %d has been authenticated", dootId)

		if newComerConfig.ActionNotify {
			return j.actionSendMessage(botReplyMsg.AuthOKMessage, &BotForceReply{true, true})
		}		
	case <-time.After(time.Duration(newComerConfig.AuthTimeout) * time.Second):
		kicked, err := j.actionKickChatMember()
		if kicked {
			delete(NewComers, newComer.Id)
			log.Printf("[!] Newcomer %d(@%s) has been kicked", newComer.Id, newComer.Username)
		}
		return kicked, err
	}

	return true, nil
}

func JobNewChatMemberWaiter(j *Job) (bool, error) {
	authMsg := j.app.Features.NewcomerQuestionnare.I18n[j.app.Lang].AuthMessage

	// will check every message if its from a newcomer to whitelist the doot, writing to the global unbuffered channel
	if _, ok := NewComers[j.ingressBody.Message.From.Id]; ok && j.ingressBody.Message.Text == authMsg {
		chNewcomer <-j.ingressBody.Message.From.Id
	}
	return true, nil
}

func (j *Job) actionSendMessage(text string, reply interface{}) (bool, error) {
	botEgressReq := &BotEgressSendMessage{
		ChatId:					j.ingressBody.Message.Chat.Id,
		Text:					text,
		ParseMode:				ParseModeMarkdown,
		DisableWebPagePreview:	true,
		DisableNotification:	true,
		ReplyToMessageId:		j.ingressBody.Message.MessageId,
		ReplyMarkup:			reply,
	}

	return botEgressReq.EgressSendToTelegram(j.app)
}

func (j *Job) actionKickChatMember() (bool, error) {
	t := time.Now().Add(time.Duration(j.app.Features.NewcomerQuestionnare.KickBanTimeout) * time.Second).Unix()

	log.Printf("[+] Kicking a newcomer %d(@%s) until %d", j.ingressBody.Message.NewChatMember.Id, j.ingressBody.Message.NewChatMember.Username, t)

	botEgressReq := &BotEgressKickChatMember{
		ChatId: j.ingressBody.Message.Chat.Id,
		UserId: j.ingressBody.Message.NewChatMember.Id,
		UntilDate: t,
	}

	return botEgressReq.EgressKickChatMember(j.app)
}