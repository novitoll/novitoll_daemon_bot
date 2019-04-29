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
	chNewcomer = make(chan int)
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

	// keyboard for authentication with the only button
	// TODO: Do the callback instead of pasting plain-text
	keyBtns := [][]KeyboardBtn{
		[]KeyboardBtn{
			KeyboardBtn{fmt.Sprintf("%s. %s - %s",
				req.AuthMessage, req.AuthPasswd, pass)},
		},
	}

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
	NewComersAuthPending[msg.Chat.Id][msg.NewChatMember.Id] = pass

	// record a newcomer and wait for his reply on the channel,
	// otherwise kick that not-doot and delete the record from this map
	j.app.Logger.WithFields(logrus.Fields{
		"chat":     msg.Chat.Id,
		"id":       msg.NewChatMember.Id,
		"username": msg.NewChatMember.Username,
	}).Warn("New member has been detected")

	// sends the welcome authentication message
	go j.SendMessageWCleanup(welcomeMsg, newComerCfg.AuthTimeout,
		&ReplyKeyboardMarkup{
			Keyboard:        keyBtns,
			OneTimeKeyboard: true,
			Selective:       true,
		})

	// blocks the current Job goroutine until either of
	// these 2 channels receive the value
	select {
	case dootId := <-chNewcomer:
		// remove from pending  map the authenticated newcomer
		delete(NewComersAuthPending[msg.Chat.Id], dootId)

		// add the authenticated user to redis's verified map
		k := fmt.Sprintf("%s-%d-%d", REDIS_USER_VERIFIED, msg.Chat.Id, dootId)
		// +10 sec, so that cronjob computing newcomers count
		// could finish in time with EVERY_LAST_SEC_7TH_DAY
		j.SaveInRedis(redisConn, k, t0, EVERY_LAST_SEC_7TH_DAY+10)

		j.app.Logger.WithFields(logrus.Fields{
			"chat": msg.Chat.Id,
			"id":   dootId,
		}).Info("Newcomer has been authenticated")

		if newComerCfg.ActionNotify {
			forceDeletion <- true
			return j.SendMessageWCleanup(req.AuthOKMessage,
				TIME_TO_DELETE_REPLY_MSG,
				&BotForceReply{
					ForceReply: false,
					Selective:  true,
				})
		} else {
			return true, nil
		}
	case <-time.After(time.Duration(newComerCfg.AuthTimeout) * time.Second):
		resp, err := j.KickChatMember()
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
				"":         msg.Chat.Id,
				"id":       msg.NewChatMember.Id,
				"username": msg.NewChatMember.Username,
			}).Warn("Newcomer has been kicked")
		}
		return resp, err
	}
}

func JobNewChatMemberAuth(j *Job) (interface{}, error) {
	/*will check every message if its from a newcomer to whitelist the doot,
	writing to the global unbuffered channel
	*/
	var isAuthMsg bool
	i18n := j.app.Features.NewcomerQuestionnare.I18n[j.app.Lang]
	msg := j.req.Message

	// ignore own and content-less messages
	if msg.From.Username == TELEGRAM_BOT_USERNAME || !j.HasMessageContent() {
		return nil, nil
	}

	// this should not compile per each message
	if authRgxp == nil {
		authRgxp = regexp.MustCompile(fmt.Sprintf("^%s", i18n.AuthMessage))
	}

	// 1. Let's check if this is for newmember auth related message or not
	matched := authRgxp.FindAllString(msg.Text, -1)
	isAuthMsg = len(matched) > 0

	// 2. ok, but let's check if user is in our auth pending map or not
	pass, isPending := NewComersAuthPending[msg.Chat.Id][msg.From.Id]

	// 2.1 pending users can not send messages except auth
	if isPending && !isAuthMsg {
		go j.DeleteMessage(&msg)
		return nil, nil
	}

	if !isAuthMsg {
		return nil, nil
	}

	// 3. ok, let's check then if user's password is legit with outs
	passOrig := fmt.Sprintf("%s. %s - %s", i18n.AuthMessage, i18n.AuthPasswd, pass)
	if isPending && passOrig == msg.Text {
		go j.onDeleteMessage(&msg, TIME_TO_DELETE_REPLY_MSG)
		chNewcomer <- msg.From.Id
	} else {
		// answer if the user has cached message (seems, a bug for desktop users)
		// or if user is already verified, he/she will get the reply
		// or if user is in auth pending but failed with password,
		// then time.After channel about will kick the fuck out that guy.
		_, err := j.SendMessageWCleanup(i18n.AuthMessageCached,
			TIME_TO_DELETE_REPLY_MSG+10,
			&BotForceReply{
				ForceReply: false,
				Selective:  true,
			})

		// delete user's message with delay
		go j.onDeleteMessage(&msg, TIME_TO_DELETE_REPLY_MSG)

		return nil, err
	}

	return true, nil
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
