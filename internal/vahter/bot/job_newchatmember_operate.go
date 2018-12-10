// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	redisClient "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
)

var (
	KickedNewcomers = []int{}
	forceDeletion = make(chan bool)
	chNewcomer    = make(chan int) // unbuffered chhanel to wait for the certain time for the newcomer's response
)

const (
	REDIS_NEWCOMER_TTL = 172800  // 2 days in sec
	NEWCOMERS_COUNT_PERIOD = 180
)

func init() {
	rand.Seed(time.Now().Unix())
}

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
	j.app.Logger.WithFields(logrus.Fields{
		"id":       newComer.Id,
		"username": newComer.Username,
	}).Warn("New member has been detected")

	// record the newcomer
	err := redisClient.SetRedisObj(string(newComer.Id), time.Now(), REDIS_NEWCOMER_TTL)
	if err != nil  {
		return nil, err
	}

	// sends the welcome authentication message
	go j.actionSendMessage(welcomeMsg, newComerConfig.AuthTimeout, &ReplyKeyboardMarkup{
		Keyboard:        keyBtns,
		OneTimeKeyboard: true,
		Selective:       true,
	})

	// blocks the current Job goroutine until either of these 2 channels receive the value
	select {
	case dootId := <-chNewcomer:
		err = redisClient.DeleteRedisObj(string(dootId))
		if err != nil {
			return nil, err
		}

		j.app.Logger.WithFields(logrus.Fields{
			"id": dootId,
		}).Info("Newcomer has been authenticated")

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
		// init the newcomers' count within N seconds
		KickedNewcomers = append(KickedNewcomers, newComer.Id)
		go j.actionCountKickedNewComers()

		response, err2 := j.actionKickChatMember()
		if err2 == nil {
			// delete the "User joined the group" event
			go j.actionDeleteMessage(&j.ingressBody.Message, TIME_TO_DELETE_REPLY_MSG)

			err = redisClient.DeleteRedisObj(string(newComer.Id))
			if err != nil {
				return nil, err
			}

			j.app.Logger.WithFields(logrus.Fields{
				"id":       newComer.Id,
				"username": newComer.Username,
			}).Warn("Newcomer has been kicked")
		}
		return response, err2
	}
}

func JobNewChatMemberAuth(j *Job) (interface{}, error) {
	authMsg := j.app.Features.NewcomerQuestionnare.I18n[j.app.Lang].AuthMessage

	// will check every message if its from a newcomer to whitelist the doot, writing to the global unbuffered channel
	if strings.ToLower(j.ingressBody.Message.Text) == strings.ToLower(authMsg) {
		newcomer, err := redisClient.GetRedisObj(string(j.ingressBody.Message.From.Id))
		if err != nil {
			return nil, err
		}
		
		if newcomer != nil {			
			go j.actionDeleteMessage(&j.ingressBody.Message, TIME_TO_DELETE_REPLY_MSG)
			chNewcomer <- j.ingressBody.Message.From.Id
		} else {
			trollReplies := j.app.Features.Administration.I18n[j.app.Lang].TrollReply
			text := trollReplies[rand.Intn(len(trollReplies))]
			return j.actionSendMessage(text, TIME_TO_DELETE_REPLY_MSG+20, &BotForceReply{
				ForceReply: false,
				Selective:  true,
			})
		}
	}
	return true, nil
}

/*
	Action functions
*/

func (j *Job) actionCountKickedNewComers() (interface{}, error) {
	select {
	case <-time.After(time.Duration(NEWCOMERS_COUNT_PERIOD) * time.Second):
		text := fmt.Sprintf(j.app.Features.NewcomerQuestionnare.I18n[j.app.Lang].StatsMessage,
		 NEWCOMERS_COUNT_PERIOD, len(KickedNewcomers))
		
		botEgressReq := &BotEgressSendMessage{
			ChatId:                j.ingressBody.Message.Chat.Id,
			Text:                  text,
			ParseMode:             ParseModeMarkdown,
			DisableWebPagePreview: true,
			DisableNotification:   true,
			ReplyMarkup:           nil,
		}

		KickedNewcomers = []int{}

		return botEgressReq.EgressSendToTelegram(j.app)
	}
}

func (j *Job) actionSendMessage(text string, deleteAfterTime int, reply interface{}) (interface{}, error) {
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

	j.app.Logger.WithFields(logrus.Fields{
		"id":       j.ingressBody.Message.NewChatMember.Id,
		"username": j.ingressBody.Message.NewChatMember.Username,
		"until":    t,
	}).Warn("Kicking a newcomer")

	botEgressReq := &BotEgressKickChatMember{
		ChatId:    j.ingressBody.Message.Chat.Id,
		UserId:    j.ingressBody.Message.NewChatMember.Id,
		UntilDate: t,
	}
	return botEgressReq.EgressKickChatMember(j.app)
}

func (j *Job) actionDeleteMessage(response *BotIngressRequestMessage, deleteAfterTime int) (interface{}, error) {
	// dirty hack to do the same function on either channel (fan-in pattern)
	select {
	case <-forceDeletion:
		return j.DeleteMessage(response)
	case <-time.After(time.Duration(deleteAfterTime) * time.Second):
		return j.DeleteMessage(response)
	}
}
