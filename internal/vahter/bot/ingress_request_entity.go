package bot

/*
	Telegram WebHook request body

	Example: URL / Text
	{
	  "message": {
	    "from": {
	      "username": "novitoll",
	      "first_name": "novitoll",
	      "is_bot": false,
	      "id": 345019684,
	      "language_code": "en-US"
	    },
	    "text": "https://weproject.kz/articles/detail/o-tom-kak-zarabotat-4000-dollarov-za-12-dney-i-ne-sidet-v-ofise/",
	    "entities": [
	      {
	        "length": 101,
	        "type": "url",
	        "offset": 0
	      }
	    ],
	    "chat": {
	      "username": "novitoll",
	      "first_name": "novitoll",
	      "type": "private",
	      "id": 345019684
	    },
	    "date": 1537020424,
	    "message_id": 28
	  },
	  "update_id": 776799951
	}
*/

type BotIngressRequest struct {
	Update_Id	uint32 `json:"update_id"`
	Message		BotIngressRequestMessage
}

type BotIngressRequestMessage struct {
	From 		User
	Text		string `json:"text"`
	Entities	[]Message
	Date		int64 `json:"date"` // time.Unix()
	MessageId	uint32 `json:"message_id"`
	Chat		Chat `json:"chat"`
	NewComer	User `json:"new_chat_members"`
}

// https://core.telegram.org/bots/api#messageentity
type Message struct {
	Type		string `json:"type"`
	Length		int `json:"length"`
	Url		string `json:"url"`
	MentionedUser	User `json:"user"`
}

// https://core.telegram.org/bots/api#user
type User struct {
	Username	string	`json:"username"`
	First_Name	string	`json:"first_name"`
	Is_Bot		bool	`json:"is_bot"`
	Language_Code	string	`json:"language_code"`
}

type Chat struct {
	Username	string `json:"username"`
	First_Name	string `json:"first_name"`
	Type		string `json:"type"`
	Id		uint32 `json:"id"`
}

var WHITELIST_URLS = []string{
	"google.com",
	"habr.com",
	"",
}
