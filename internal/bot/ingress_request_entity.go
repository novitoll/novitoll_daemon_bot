// SPDX-License-Identifier: GPL-2.0
package bot

type BotInResp struct {
	Ok          bool        `json:"ok"`
	Result      BotInReqMsg `json:"result"`
	ErrorCode   int         `json:"error_code"`
	Description string      `json:"description"`
}

type BotInRespMult struct {
	Ok          bool           `json:"ok"`
	Result      []*BotInReqMsg `json:"result"`
	ErrorCode   int            `json:"error_code"`
	Description string         `json:"description"`
}

type BotInResp2 struct {
	Ok     bool `json:"ok"`
	Result bool `json:"result"`
}

type BotInReq struct {
	Update_Id int `json:"update_id"`
	Message   BotInReqMsg
	CallbackQuery CallbackQuery
}

type BotInReqMsg struct {
	From                User
	Text                string `json:"text"`
	Entities            []Message
	Date                int64   `json:"date"` // time.Unix()
	MessageId           int     `json:"message_id"`
	Chat                Chat    `json:"chat"`
	NewChatMembers      []User  `json:"new_chat_members"`
	NewChatMember       User    `json:"new_chat_member"`
	NewChatParticipant  User    `json:"new_chat_participant"`
	LeftChatParticipant User    `json:"left_chat_participant"`
	Sticker             Sticker `json:"sticker"`
	Caption             string  `json:"caption"`
}

// https://core.telegram.org/bots/api#user
type User struct {
	Id           int    `json:"id"`
	Username     string `json:"username"`
	First_Name   string `json:"first_name"`
	IsBot        bool   `json:"is_bot"`
	LanguageCode string `json:"language_code"`
}

// https://core.telegram.org/bots/api#messageentity
type Message struct {
	Type          string `json:"type"`
	Length        int    `json:"length"`
	Url           string `json:"url"`
	MentionedUser User   `json:"user"`
}

type Chat struct {
	Username   string `json:"username"`
	First_Name string `json:"first_name"`
	Type       string `json:"type"`
	Id         int    `json:"id"`
	Title      string `json:"title"`
}

// https://core.telegram.org/bots/api#sticker
type Sticker struct {
	FileId string `json:"file_id"`
}

// https://core.telegram.org/bots/api#callbackquery
type CallbackQuery struct {
	Id 	string `json:"id"`
	InlineMessageId string `json:"inline_message_id"`
	From 	User `json:"from"`
	Message 	BotInReqMsg `json:"message"`
	ChatInstance string `json:"chat_instance"`
}
