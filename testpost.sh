#!/bin/bash

set -e 

curl -X POST -H 'Content-Type: application/json' -d '{"message":{"from":{"username":"novitoll","first_name":"novitoll","is_bot":false,"id":345019684,"language_code":"en-US"},"text":"asdasdhttps://weproject.kz/articles/detail/o-tom-kak-zarabotat-4000-dollarov-za-12-dney-i-ne-sidet-v-ofise/","entities":[{"length":101,"type":"url","offset":0}],"chat":{"username":"novitoll","first_name":"novitoll","type":"private","id":345019684},"date":1537020424,"message_id":28},"update_id":776799951}' http://localhost:8080/check