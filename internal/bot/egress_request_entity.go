// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"os"
)

/*
	Telegram Outbound (sendMessage command) request body

	Example: URL / Text
	{
	  "chat_id": 345019684,
	  "text": "Bot reply",
	  "parse_mode": "Markdown",
	  "disable_web_page_preview": true,
	  "disable_notification": false,
	  "reply_to_message_id": 28
	}
*/

var (
	GET               = "GET"
	POST              = "POST"
	TELEGRAM_URL      = "https://api.telegram.org/bot%s/%s"
	TELEGRAM_TOKEN    = "123"
	ParseModeMarkdown = "Markdown"
	ParseModeHTML     = "HTML"
)

func init() {
	if u, ok := os.LookupEnv("TELEGRAM_URL"); ok {
		TELEGRAM_URL = u
	}
	if t, ok := os.LookupEnv("TELEGRAM_TOKEN"); ok {
		TELEGRAM_TOKEN = t
	}
}

// https://core.telegram.org/bots/api#sendmessage
type BotSendMsg struct {
	ChatId           int         `json:"chat_id"`
	Text             string      `json:"text"`
	ParseMode        string      `json:"parse_mode"`
	ReplyToMessageId int         `json:"reply_to_message_id"`
	ReplyMarkup      interface{} `json:"reply_markup"`
}

// https://core.telegram.org/bots/api#forcereply
type BotForceReply struct {
	ForceReply bool `json:"force_reply"`
	Selective  bool `json:"selective"`
}

// https://core.telegram.org/bots/api#replykeyboardmarkup
type ReplyKeyboardMarkup struct {
	Keyboard        [][]KeyboardBtn `json:"keyboard"`
	OneTimeKeyboard bool            `json:"one_time_keyboard"`
	Selective       bool            `json:"selective"`
}

// https://core.telegram.org/bots/api#inlinekeyboardmarkup
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// https://core.telegram.org/bots/api#inlinekeyboardbutton
type InlineKeyboardButton struct {
	Text string `json:"text"`
	CallbackData string `json:"callback_data"`
}

// https://core.telegram.org/bots/api#keyboardbutton
type KeyboardBtn struct {
	Text string `json:"text"`
}

// https://core.telegram.org/bots/api#kickchatmember
type BotKickChatMember struct {
	ChatId    int   `json:"chat_id"`
	UserId    int   `json:"user_id"`
	UntilDate int64 `json:"until_date"`
}

// https://core.telegram.org/bots/api#deletemessage
type BotDeleteMsg struct {
	ChatId    int `json:"chat_id"`
	MessageId int `json:"message_id"`
}

// https://core.telegram.org/bots/api#getchatadministrators
type BotGetAdmins struct {
	ChatId int `json:"chat_id"`
}

// https://core.telegram.org/bots/api#answercallbackquery
type BotAnswerCallbackQuery struct {
	CallbackQueryId string `json:"callback_query_id"`	
}
