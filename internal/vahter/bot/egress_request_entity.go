package bot

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
	GET = "GET"
	POST = "POST"
	TELEGRAM_URL = "http://telegrammock:8081/%s"
	ParseModeMarkdown = "Markdown"
	ParseModeHTML = "HTML"
)

// https://core.telegram.org/bots/api#sendmessage
type TelegramRequestBody struct {
	ChatId			uint32 `json:"chat_id"`
	Text			string `json:"text"`
	ParseMode		string `json:"parse_mode"`
	DisableWebPagePreview	bool `json:"disable_web_page_preview"`
	DisableNotification	bool `json:"disable_notification"`
	ReplyToMessageId	uint32 `json:"reply_to_message_id"`
}
