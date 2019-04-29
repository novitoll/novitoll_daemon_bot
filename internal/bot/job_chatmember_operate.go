// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"regexp"
	"time"

	redis "github.com/novitoll/novitoll_daemon_bot/pkg/redis_client"
	"github.com/novitoll/novitoll_daemon_bot/pkg/utils"
	"github.com/sirupsen/logrus"
)

var (
	authRgxp      *regexp.Regexp
	forceDeletion = make(chan bool)
	// unbuffered chanel to wait for the certain time
	// for the newcomer's response.
	chNewcomer = make(chan []int)
	// we could store this map in Redis as well,
	// but once we have the record here, we have to
	// check Redis (open TCP connection) per each message
	// because we don't know beforehand if the message is
	// from the Auth pending user or not. So keep in memory
	// a double nested hashmap for multiple chats.
	// {
	//   <chat_id>: {
	//   	<user_id>: <timestamp>,
	//   	<user_id>: <timestamp>,
	//	},
	//   <chat_id>: {..}
	// }
	NewComersAuthPending = make(map[int]map[int]string)
)

func JobNewChatMemberDetector(j *Job) (interface{}, error) {
	// for short code reference
	msg := j.req.Message
	newComerCfg := j.app.Features.NewcomerQuestionnare
	req := newComerCfg.I18n[j.app.Lang]

	// do not validate yourself
	if !newComerCfg.Enabled || msg.NewChatMember.Id == 0 ||
		msg.NewChatMember.Username == TELEGRAM_BOT_USERNAME {
		return false, nil
	}

	// init a randomized auth check
	pass := utils.RandStringRunes(9)
	auth := fmt.Sprintf("%s. %s - %s", req.AuthMessage, req.AuthPasswd, pass)

	welcomeMsg := fmt.Sprintf(req.WelcomeMessage,
		newComerCfg.AuthTimeout, newComerCfg.KickBanTimeout)

	redisConn := redis.GetRedisConnection()
	defer redisConn.Close()

	t0 := time.Now().Unix()

	if _, ok := NewComersAuthPending[msg.Chat.Id]; !ok {
		// init inner map if it's first time for this chat
		NewComersAuthPending[msg.Chat.Id] = map[int]string{}
	}
	// store a newcomer per chat
	if _, isPending := NewComersAuthPending[msg.Chat.Id][msg.NewChatMember.Id]; isPending {
		j.app.Logger.Info("Deleting pending user message")
		go j.DeleteMessage(&msg)
		return nil, nil
	}

	NewComersAuthPending[msg.Chat.Id][msg.NewChatMember.Id] = pass

	// record a newcomer and wait for his reply on the channel,
	// otherwise kick that not-doot and delete the record from this map
	j.app.Logger.WithFields(logrus.Fields{
		"chat":     msg.Chat.Id,
		"id":       msg.NewChatMember.Id,
		"username": msg.NewChatMember.Username,
	}).Warn("New member has been detected")

	keyBtns := [][]InlineKeyboardButton{
		[]InlineKeyboardButton{
			InlineKeyboardButton{Text: auth, CallbackData: pass},
		},
	}

	// sends the welcome authentication message with a callback
	// After user hits the button, CallbackQuery will be sent back
	// This will eliminate bots sending password in plain text by reading it.
	// kudos for the idea to @kazgeek
	go j.SendMessageWReply(welcomeMsg, msg.MessageId,
		&InlineKeyboardMarkup{
			InlineKeyboard: keyBtns,
		})

	// blocks the current Job goroutine until either of
	// these 2 channels receive the value
	select {
	case doot := <-chNewcomer:
		dootId, dootChatId := doot[0], doot[1]
		// remove from pending  map the authenticated newcomer
		delete(NewComersAuthPending[dootChatId], dootId)

		// add the authenticated user to redis's verified map
		k := fmt.Sprintf("%s-%d-%d", REDIS_USER_VERIFIED, dootChatId, dootId)
		// +10 sec, so that cronjob computing newcomers count
		// could finish in time with EVERY_LAST_SEC_7TH_DAY
		j.SaveInRedis(redisConn, k, t0, EVERY_LAST_SEC_7TH_DAY+10)

		j.app.Logger.WithFields(logrus.Fields{
			"chat": dootChatId,
			"id":   dootId,
		}).Info("Newcomer has been authenticated")

		if newComerCfg.ActionNotify {
			forceDeletion <- true
			return j.SendMessageWCleanup(dootChatId, req.AuthOKMessage,
				TIME_TO_DELETE_REPLY_MSG,
				&BotForceReply{
					ForceReply: false,
					Selective:  true,
				})
		} else {
			return true, nil
		}
	case <-time.After(time.Duration(newComerCfg.AuthTimeout) * time.Second):
		resp, err := j.KickChatMember(msg.NewChatMember.Id, msg.NewChatMember.Username)
		if err == nil {
			// delete the "User joined the group" event
			go j.onDeleteMessage(&msg, TIME_TO_DELETE_REPLY_MSG)

			// delete un-authenticated user from pending map
			delete(NewComersAuthPending[msg.Chat.Id], msg.NewChatMember.Id)

			// record this event in redis's kicked users map
			k := fmt.Sprintf("%s-%d-%d", REDIS_USER_KICKED,
				msg.Chat.Id, msg.NewChatMember.Id)

			// same +10 sec as for REDIS_USER_VERIFIED
			j.SaveInRedis(redisConn, k, t0, EVERY_LAST_SEC_7TH_DAY+10)

			j.app.Logger.WithFields(logrus.Fields{
				"chat":     msg.Chat.Id,
				"id":       msg.NewChatMember.Id,
				"username": msg.NewChatMember.Username,
			}).Warn("Newcomer has been kicked")
		}
		return resp, err
	}
}

func JobNewChatMemberAuth(j *Job) (interface{}, error) {
	// will check CallbackQuery only from a newcomer to whitelist the doot,
	// writing to the global unbuffered channel
	cb := j.req.CallbackQuery

	if cb.Id == "" {
		return nil, nil
	}

	req := &BotAnswerCallbackQuery{
		CallbackQueryId: cb.Id,
	}

	origPass, isPending := NewComersAuthPending[cb.Message.Chat.Id][cb.From.Id]

	if isPending {
		if origPass == cb.Data {
			doot := []int{cb.From.Id, cb.Message.Chat.Id}
			chNewcomer <- doot
		} else {
			j.app.Logger.Info("Pending user's callback password was incorrect. That's weird")
		}
	} else {
		j.app.Logger.Info("Not pending user clicked the button")
		// TODO: j.SendMessage(..)
	}

	return req.AnswerCallbackQuery(j.app)
}

func JobLeftParticipantDetector(j *Job) (interface{}, error) {
	msg := j.req.Message
	left := msg.LeftChatParticipant

	if left.Id == 0 {
		return false, nil
	}

	redisConn := redis.GetRedisConnection()
	defer redisConn.Close()

	k := fmt.Sprintf("%s-%d-%d", REDIS_USER_LEFT, msg.Chat.Id, left.Id)
	t0 := time.Now()
	j.SaveInRedis(redisConn, k, t0, EVERY_LAST_SEC_7TH_DAY+10)
	return nil, nil
}

/*
	Action functions
*/

func (j *Job) onDeleteMessage(resp *BotInReqMsg, delay uint8) (interface{}, error) {
	// dirty hack to do the same function on either channel (fan-in pattern)
	select {
	case <-forceDeletion:
		return j.DeleteMessage(resp)
	case <-time.After(time.Duration(delay) * time.Second):
		return j.DeleteMessage(resp)
	}
}
