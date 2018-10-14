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

	Example: 2
	{
		"update_id": 53205695,
		"message": {
			"message_id": 105,
			"from": {
			  "id": 345019684,
			  "is_bot": false,
			  "first_name": "novitoll",
			  "username": "novitoll",
			  "language_code": "en-US"
			},
			"chat": {
			  "id": -253761934,
			  "title": "test_novitoll_daemon_bot",
			  "type": "group",
			  "all_members_are_administrators": true
			},
			"date": 1539514928,
			"new_chat_participant": {
			  "id": 553713145,
			  "is_bot": true,
			  "first_name": "novitoll_daemon_bot",
			  "username": "novitoll_daemon_bot"
			},
			"new_chat_member": {
			  "id": 553713145,
			  "is_bot": true,
			  "first_name": "novitoll_daemon_bot",
			  "username": "novitoll_daemon_bot"
			},
			"new_chat_members": [
			  {
			    "id": 553713145,
			    "is_bot": true,
			    "first_name": "novitoll_daemon_bot",
			    "username": "novitoll_daemon_bot"
			  }
			]
		}
	}

*/

type BotIngressRequest struct {
	Update_Id	int `json:"update_id"`
	Message		BotIngressRequestMessage
}

type BotIngressRequestMessage struct {
	From 		User
	Text		string `json:"text"`
	Entities	[]Message
	Date		int64 `json:"date"` // time.Unix()
	MessageId	int `json:"message_id"`
	Chat		Chat `json:"chat"`
	NewChatMembers	[]User `json:"new_chat_members"`
	NewChatMember User `json:"new_chat_member"`
	NewChatParticipant User `json:"new_chat_participant"`
}

// https://core.telegram.org/bots/api#user
type User struct {
	Id 			int `json:"id"`
	Username	string	`json:"username"`
	First_Name	string	`json:"first_name"`
	IsBot		bool	`json:"is_bot"`
	LanguageCode	string	`json:"language_code"`
}

// https://core.telegram.org/bots/api#messageentity
type Message struct {
	Type		string `json:"type"`
	Length		int `json:"length"`
	Url		string `json:"url"`
	MentionedUser	User `json:"user"`
}

type Chat struct {
	Username	string `json:"username"`
	First_Name	string `json:"first_name"`
	Type		string `json:"type"`
	Id		int `json:"id"`
	Title 	string `json:"title"`
}

var WHITELIST_URLS = []string{
	"google.com",
	"habr.com",
	"",
}
