// SPDX-License-Identifier: GPL-2.0
package bot

const (
	BDFL                     = "novitoll"
	TELEGRAM_BOT_USERNAME    = "novitoll_daemon_bot"
	TIME_TO_DELETE_REPLY_MSG = 10
)

var (
	NewComersAuthPending  = make(map[int]interface{})
	NewComersAuthVerified = make(map[int]interface{})
	NewComersKicked       = make(map[int]interface{})
	// Map to store user message statistics.
	// Data in the map is cleaned up when the CronJob executes (every last second of 7th day)
	UserStatistics = make(map[int]*UserMessageStats)
)
