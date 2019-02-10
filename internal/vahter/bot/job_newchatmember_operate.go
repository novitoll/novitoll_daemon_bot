// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	forceDeletion = make(chan bool)
	// unbuffered chhanel to wait for the certain time
	// for the newcomer's resp
	chNewcomer    = make(chan int)
)

func JobNewChatMemberDetector(j *Job) (interface{}, error) {
	// for short code reference
	newComer := j.req.Message.NewChatMember
	newComerConfig := j.app.Features.NewcomerQuestionnare
	req := newComerConfig.I18n[j.app.Lang]

	if !newComerConfig.Enabled || newComer.Id == 0 || 
		newComer.Username == TELEGRAM_BOT_USERNAME {
		return false, nil
	}

	// init vars
	keyBtns := [][]KeyboardBtn{
		[]KeyboardBtn{
			KeyboardBtn{req.AuthMessage},
		},
	}
	welcomeMsg := fmt.Sprintf(req.WelcomeMessage, 
		newComerConfig.AuthTimeout, newComerConfig.KickBanTimeout)

	t0 := time.Now()
	NewComersAuthPending[newComer.Id] = t0

	// record a newcomer and wait for his reply on the channel,
	// otherwise kick that not-doot and delete the record from this map
	j.app.Logger.WithFields(logrus.Fields{
		"id":       newComer.Id,
		"username": newComer.Username,
	}).Warn("New member has been detected")

	// sends the welcome authentication message
	go j.onSendMessage(welcomeMsg, newComerConfig.AuthTimeout,
		&ReplyKeyboardMarkup{
			Keyboard:        keyBtns,
			OneTimeKeyboard: true,
			Selective:       true,
		})

	// blocks the current Job goroutine until either of 
	// these 2 channels receive the value
	select {
	case dootId := <-chNewcomer:
		delete(NewComersAuthPending, dootId)
		NewComersAuthVerified[dootId] = t0

		j.app.Logger.WithFields(logrus.Fields{
			"id": dootId,
		}).Info("Newcomer has been authenticated")

		if newComerConfig.ActionNotify {
			forceDeletion <- true
			return j.onSendMessage(req.AuthOKMessage,
					TIME_TO_DELETE_REPLY_MSG,
					&BotForceReply{
						ForceReply: false,
						Selective:  true,
					})
		} else {
			return true, nil
		}
	case <-time.After(time.Duration(newComerConfig.AuthTimeout) * time.Second):
		resp, err := j.onKickChatMember()
		if err == nil {

			// delete the "User joined the group" event
			go j.onDeleteMessage(&j.req.Message, TIME_TO_DELETE_REPLY_MSG)

			delete(NewComersAuthPending, newComer.Id)
			NewComersKicked[newComer.Id] = t0

			j.app.Logger.WithFields(logrus.Fields{
				"id":       newComer.Id,
				"username": newComer.Username,
			}).Warn("Newcomer has been kicked")
		}
		return resp, err
	}
}

func JobNewChatMemberAuth(j *Job) (interface{}, error) {
	i18n := j.app.Features.NewcomerQuestionnare.I18n[j.app.Lang]

	// will check every message if its from a newcomer to whitelist the doot,
	// writing to the global unbuffered channel
	if strings.ToLower(j.req.Message.Text) == strings.ToLower(i18n.AuthMessage) {
		
		// doot is verified
		if _, ok := NewComersAuthPending[j.req.Message.From.Id]; ok {
			go j.onDeleteMessage(&j.req.Message, TIME_TO_DELETE_REPLY_MSG)
			chNewcomer <- j.req.Message.From.Id
		
		// answer if the user has cached message (seems, a bug for desktop users)
		} else {
			_, err := j.onSendMessage(i18n.AuthMessageCached, 
				TIME_TO_DELETE_REPLY_MSG + 10,
				 &BotForceReply{
					ForceReply: false,
					Selective:  true,
				})
			
			// delete user's message with delay
			go j.onDeleteMessage(&j.req.Message, TIME_TO_DELETE_REPLY_MSG)

			return nil, err
		}
	}
	return true, nil
}

/*
	Action functions
*/

func (j *Job) onSendMessage(text string, delay uint8, reply interface{}) (interface{}, error) {
	botEgressReq := &BotSendMsg{
		ChatId:                j.req.Message.Chat.Id,
		Text:                  text,
		ParseMode:             ParseModeMarkdown,
		ReplyToMessageId:      j.req.Message.MessageId,
		ReplyMarkup:           reply,
	}
	replyMsgBody, err := botEgressReq.SendMsg(j.app)
	if err != nil {
		return false, err
	}

	if replyMsgBody != nil {
		// cleanup reply messages
		go j.onDeleteMessage(replyMsgBody, delay)
	}

	return replyMsgBody, err
}

func (j *Job) onKickChatMember() (interface{}, error) {
	t := time.Now().Add(time.Duration(j.app.Features.
		NewcomerQuestionnare.KickBanTimeout) * time.Second).Unix()

	j.app.Logger.WithFields(logrus.Fields{
		"id":       j.req.Message.NewChatMember.Id,
		"username": j.req.Message.NewChatMember.Username,
		"until":    t,
	}).Warn("Kicking a newcomer")

	botEgressReq := &BotKickChatMember{
		ChatId:    j.req.Message.Chat.Id,
		UserId:    j.req.Message.NewChatMember.Id,
		UntilDate: t,
	}
	return botEgressReq.KickChatMember(j.app)
}

func (j *Job) onDeleteMessage(resp *BotInReqMsg, delay uint8) (interface{}, error) {
	// dirty hack to do the same function on either channel (fan-in pattern)
	select {
	case <-forceDeletion:
		return j.DeleteMessage(resp)
	case <-time.After(time.Duration(delay) * time.Second):
		return j.DeleteMessage(resp)
	}
}