// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/novitoll/novitoll_daemon_bot/internal/utils"
	redis "github.com/novitoll/novitoll_daemon_bot/internal/vahter/redis_client"
	"github.com/sirupsen/logrus"
)

var (
	forceDeletion = make(chan bool)
	// unbuffered chhanel to wait for the certain time
	// for the newcomer's resp
	chNewcomer = make(chan int)
)

func JobNewChatMemberDetector(j *Job) (interface{}, error) {
	// for short code reference
	newComer := j.req.Message.NewChatMember
	newComerCfg := j.app.Features.NewcomerQuestionnare
	req := newComerCfg.I18n[j.app.Lang]

	if !newComerCfg.Enabled || newComer.Id == 0 ||
		newComer.Username == TELEGRAM_BOT_USERNAME {
		return false, nil
	}

	p := utils.RandStringRunes(6)

	// init vars
	keyBtns := [][]KeyboardBtn{
		[]KeyboardBtn{
			KeyboardBtn{fmt.Sprintf(req.AuthMessage, p)},
		},
	}
	welcomeMsg := fmt.Sprintf(req.WelcomeMessage,
		newComerCfg.AuthTimeout, newComerCfg.KickBanTimeout)

	redisConn := redis.GetRedisConnection()
	defer redisConn.Close()

	t0 := time.Now()
	k := fmt.Sprintf("%s-%d", REDIS_USER_PENDING, newComer.Id)

	j.SaveInRedis(redisConn, k, p, int(newComerCfg.AuthTimeout+10))

	// record a newcomer and wait for his reply on the channel,
	// otherwise kick that not-doot and delete the record from this map
	j.app.Logger.WithFields(logrus.Fields{
		"id":       newComer.Id,
		"username": newComer.Username,
	}).Warn("New member has been detected")

	// sends the welcome authentication message
	go j.onSendMessage(welcomeMsg, newComerCfg.AuthTimeout,
		&ReplyKeyboardMarkup{
			Keyboard:        keyBtns,
			OneTimeKeyboard: true,
			Selective:       true,
		})

	// blocks the current Job goroutine until either of
	// these 2 channels receive the value
	select {
	case dootId := <-chNewcomer:
		k = fmt.Sprintf("%s-%d", REDIS_USER_VERIFIED, dootId)
		// +10 sec, so that cronjob computing newcomers count
		// could finish in time with EVERY_LAST_SEC_7TH_DAY
		j.SaveInRedis(redisConn, k, t0, EVERY_LAST_SEC_7TH_DAY+10)

		j.app.Logger.WithFields(logrus.Fields{
			"id": dootId,
		}).Info("Newcomer has been authenticated")

		if newComerCfg.ActionNotify {
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
	case <-time.After(time.Duration(newComerCfg.AuthTimeout) * time.Second):
		resp, err := j.onKickChatMember()
		if err == nil {

			// delete the "User joined the group" event
			go j.onDeleteMessage(&j.req.Message, TIME_TO_DELETE_REPLY_MSG)

			k = fmt.Sprintf("%s-%d", REDIS_USER_KICKED, newComer.Id)
			// same +10 sec as for REDIS_USER_VERIFIED
			j.SaveInRedis(redisConn, k, t0, EVERY_LAST_SEC_7TH_DAY+10)

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
	redisConn := redis.GetRedisConnection()
	defer redisConn.Close()

	k := fmt.Sprintf("%s-%d", REDIS_USER_PENDING, j.req.Message.From.Id)
	p := j.GetFromRedis(redisConn, k)

	if strings.ToLower(j.req.Message.Text) == strings.ToLower(fmt.Sprintf(i18n.AuthMessage, p)) {
		// doot is verified
		if p != nil {
			go j.onDeleteMessage(&j.req.Message, TIME_TO_DELETE_REPLY_MSG)
			chNewcomer <- j.req.Message.From.Id

			// answer if the user has cached message (seems, a bug for desktop users)
		} else {
			_, err := j.onSendMessage(i18n.AuthMessageCached,
				TIME_TO_DELETE_REPLY_MSG+10,
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

func JobLeftParticipantDetector(j *Job) (interface{}, error) {
	left := j.req.Message.LeftChatParticipant

	if left.Id == 0 {
		return false, nil
	}

	redisConn := redis.GetRedisConnection()
	defer redisConn.Close()

	k := fmt.Sprintf("%s-%d", REDIS_USER_LEFT, left.Id)
	t0 := time.Now()
	j.SaveInRedis(redisConn, k, t0, EVERY_LAST_SEC_7TH_DAY+10)
	return nil, nil
}

/*
	Action functions
*/

func (j *Job) onSendMessage(text string, delay uint8, reply interface{}) (interface{}, error) {
	botEgressReq := &BotSendMsg{
		ChatId:           j.req.Message.Chat.Id,
		Text:             text,
		ParseMode:        ParseModeMarkdown,
		ReplyToMessageId: j.req.Message.MessageId,
		ReplyMarkup:      reply,
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
