// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"log"
	"time"
)

const (
	TIME_TO_DELETE_REPLY_MSG = 10
)

var (
	NewComers     = make(map[int]interface{})
	forceDeletion = make(chan bool)
	chNewcomer    = make(chan int) // unbuffered chhanel to wait for the certain time for the newcomer's response
)

func JobNewChatMemberDetector(j *Job) (interface{}, error) {
	// for short code reference
	newComer := j.ingressBody.Message.NewChatMember
	newComerConfig := j.app.Features.NewcomerQuestionnare
	botReplyMsg := newComerConfig.I18n[j.app.Lang]

	if !newComerConfig.Enabled || newComer.Id == 0 || newComer.Username == "@novitoll_daemon_bot" {
		return false, nil
	}

	// TODO: pointers?
	keyBtns := [][]KeyboardButton{
		[]KeyboardButton{
			KeyboardButton{botReplyMsg.AuthMessage},
		},
	}

	welcomeMsg := fmt.Sprintf(botReplyMsg.WelcomeMessage, newComerConfig.AuthTimeout, newComerConfig.KickBanTimeout)

	// record a newcomer and wait for his reply on the channel,
	// otherwise kick that not-doot and delete the record from this map
	log.Printf("[.] New member %d(@%s) has been detected", newComer.Id, newComer.Username)
	NewComers[newComer.Id] = time.Now()

	// sends the welcome authentication message
	go j.actionSendMessage(welcomeMsg, newComerConfig.AuthTimeout, &ReplyKeyboardMarkup{
		Keyboard:        keyBtns,
		OneTimeKeyboard: true,
		Selective:       true,
	})

	// blocks the current Job goroutine until either of these 2 channels receive the value
	select {
	case dootId := <-chNewcomer:
		delete(NewComers, dootId)
		log.Printf("[+] Newcomer %d has been authenticated", dootId)

		if newComerConfig.ActionNotify {
			forceDeletion <- true
			return j.actionSendMessage(botReplyMsg.AuthOKMessage, TIME_TO_DELETE_REPLY_MSG, &BotForceReply{
				ForceReply: false,
				Selective:  true,
			})
		} else {
			return true, nil
		}
	case <-time.After(time.Duration(newComerConfig.AuthTimeout) * time.Second):
		response, err := j.actionKickChatMember()
		if err == nil {

			// delete the "User joined the group" event
			go j.actionDeleteMessage(&j.ingressBody.Message, TIME_TO_DELETE_REPLY_MSG)

			delete(NewComers, newComer.Id)
			log.Printf("[!] Newcomer %d(@%s) has been kicked", newComer.Id, newComer.Username)
		}
		return response, err
	}
}

func JobNewChatMemberWaiter(j *Job) (interface{}, error) {
	authMsg := j.app.Features.NewcomerQuestionnare.I18n[j.app.Lang].AuthMessage

	// will check every message if its from a newcomer to whitelist the doot, writing to the global unbuffered channel
	if _, ok := NewComers[j.ingressBody.Message.From.Id]; ok && j.ingressBody.Message.Text == authMsg {
		go j.actionDeleteMessage(&j.ingressBody.Message, TIME_TO_DELETE_REPLY_MSG)
		chNewcomer <- j.ingressBody.Message.From.Id
	}
	return true, nil
}

/*
	Action functions
*/

func (j *Job) actionSendMessage(text string, deleteAfterTime uint8, reply interface{}) (interface{}, error) {
	botEgressReq := &BotEgressSendMessage{
		ChatId:                j.ingressBody.Message.Chat.Id,
		Text:                  text,
		ParseMode:             ParseModeMarkdown,
		DisableWebPagePreview: true,
		DisableNotification:   true,
		ReplyToMessageId:      j.ingressBody.Message.MessageId,
		ReplyMarkup:           reply,
	}
	replyMsgBody, err := botEgressReq.EgressSendToTelegram(j.app)
	if err != nil {
		return false, err
	}

	if replyMsgBody != nil {
		// cleanup reply messages
		go j.actionDeleteMessage(replyMsgBody, deleteAfterTime)
	}

	return replyMsgBody, err
}

func (j *Job) actionKickChatMember() (interface{}, error) {
	t := time.Now().Add(time.Duration(j.app.Features.NewcomerQuestionnare.KickBanTimeout) * time.Second).Unix()

	log.Printf("[.] Kicking a newcomer %d(@%s) until %d", j.ingressBody.Message.NewChatMember.Id, j.ingressBody.Message.NewChatMember.Username, t)

	botEgressReq := &BotEgressKickChatMember{
		ChatId:    j.ingressBody.Message.Chat.Id,
		UserId:    j.ingressBody.Message.NewChatMember.Id,
		UntilDate: t,
	}
	return botEgressReq.EgressKickChatMember(j.app)
}

func (j *Job) actionDeleteMessage(response *BotIngressRequestMessage, deleteAfterTime uint8) (interface{}, error) {
	// dirty hack to do the same function on either channel (fan-in pattern)
	select {
	case <-forceDeletion:
		return j.deleteMessage(response, deleteAfterTime)
	case <-time.After(time.Duration(deleteAfterTime) * time.Second):
		return j.deleteMessage(response, deleteAfterTime)
	}
}

func (j *Job) deleteMessage(response *BotIngressRequestMessage, deleteAfterTime uint8) (interface{}, error) {
	log.Printf("[.] Deleting a reply message %d", response.MessageId)

	botEgressReq := &BotEgressDeleteMessage{
		ChatId:    response.Chat.Id,
		MessageId: response.MessageId,
	}
	_, err := botEgressReq.EgressDeleteMessage(j.app)
	if err != nil {
		return false, err
	}
	return true, nil
}
