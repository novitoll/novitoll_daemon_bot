// SPDX-License-Identifier: GPL-2.0
package bot

const (
	BDFL                     = "novitoll"
	TELEGRAM_BOT_USERNAME    = "novitoll_daemon_bot"
	TIME_TO_DELETE_REPLY_MSG = 10
	// 7 days - 1 sec
	EVERY_LAST_SEC_7TH_DAY 	 = 604799
	// 15 min
	NEWCOMER_URL_POST_DELAY  = 900
)

var (
	// All these REDIS_* are Redis keys with following common
	// nested map:
	// { <chat-id>: { <user_id>: timestamp } }
	REDIS_USER_PENDING       = "NewComersAuthPending"
	REDIS_USER_VERIFIED      = "NewComersAuthVerified"
	REDIS_USER_KICKED        = "NewComersAuthKicked"
	REDIS_USER_LEFT          = "ParticipantLeft"
	REDIS_USER_PREV_LEFT     = "ParticipantLeftPrev"
	REDIS_USER_PREV_KICK     = "NewComersAuthKickedPrev"
	REDIS_USER_PREV_VERIFIED = "NewComersAuthVerifiedPrev"
)
