package bot

/*
	Example: URL / Text
{
  "message": {
    "from": {
      "username": "novitoll",
      "first_name": "novitoll",
      "is_bot": false,
      "id": 1,
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
      "id": 1
    },
    "date": 1537020424,
    "message_id": 28
  },
  "update_id": 776799951
}

	Example: Newcomer
{
  "update_id": 53205695,
  "message": {
    "message_id": 105,
    "from": {
      "id": 1,
      "is_bot": false,
      "first_name": "novitoll",
      "username": "novitoll",
      "language_code": "en-US"
    },
    "chat": {
      "id": -4,
      "title": "test_novitoll_daemon_bot",
      "type": "group",
      "all_members_are_administrators": true
    },
    "date": 1539514928,
    "new_chat_participant": {
      "id": 3,
      "is_bot": true,
      "first_name": "novitoll_daemon_bot",
      "username": "novitoll_daemon_bot"
    },
    "new_chat_member": {
      "id": 3,
      "is_bot": true,
      "first_name": "novitoll_daemon_bot",
      "username": "novitoll_daemon_bot"
    },
    "new_chat_members": [
      {
        "id": 3,
        "is_bot": true,
        "first_name": "novitoll_daemon_bot",
        "username": "novitoll_daemon_bot"
      }
    ]
  }
}

	Example: Response body from Telegram on HTTP POST methods
{
  "ok": true,
  "result": {
    "message_id": 90,
    "from": {
      "id": 645595095,
      "is_bot": true,
      "first_name": "novitoll_daemon_bot_test",
      "username": "novitoll_daemon_bot_test_bot"
    },
    "chat": {
      "id": -2,
      "title": "test2_novitoll_daemon_bot",
      "type": "group",
      "all_members_are_administrators": false
    },
    "date": 1539773176,
    "reply_to_message": {
      "message_id": 87,
      "from": {
        "id": 1,
        "is_bot": false,
        "first_name": "N2",
        "username": "Novitoll_2"
      },
      "chat": {
        "id": -2,
        "title": "test2_novitoll_daemon_bot",
        "type": "group",
        "all_members_are_administrators": false
      },
      "date": 1539773172,
      "new_chat_participant": {
        "id": 1,
        "is_bot": false,
        "first_name": "N2",
        "username": "Novitoll_2"
      },
      "new_chat_member": {
        "id": 1,
        "is_bot": false,
        "first_name": "N2",
        "username": "Novitoll_2"
      },
      "new_chat_members": [
        {
          "id": 1,
          "is_bot": false,
          "first_name": "N2",
          "username": "Novitoll_2"
        }
      ]
    },
    "text": "\\u0421\\u043f\\u0430\\u0441\\u0438\\u0431\\u043e,\\u0432\\u044b\\u0430\\u0432\\u0442\\u043e\\u0440\\u0438\\u0437\\u043e\\u0432\\u0430\\u043d\\u044b."
  }
}
*/

type BotIngressResponse struct {
	Ok          bool                     `json:"ok"`
	Result      BotIngressRequestMessage `json:"result"`
	ErrorCode   int                      `json:"error_code"`
	Description string                   `json:"description"`
}

type BotIngressResponse2 struct {
  Ok          bool `json:"ok"`
  Result      bool `json:"result"`
}

type BotIngressRequest struct {
	Update_Id int `json:"update_id"`
	Message   BotIngressRequestMessage
}

type BotIngressRequestMessage struct {
	From               User
	Text               string `json:"text"`
	Entities           []Message
	Date               int64  `json:"date"` // time.Unix()
	MessageId          int    `json:"message_id"`
	Chat               Chat   `json:"chat"`
	NewChatMembers     []User `json:"new_chat_members"`
	NewChatMember      User   `json:"new_chat_member"`
	NewChatParticipant User   `json:"new_chat_participant"`
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
