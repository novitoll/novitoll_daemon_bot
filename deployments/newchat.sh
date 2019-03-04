#!/bin/bash

set -e

function usage() {
	echo """./newchat.sh [<arguments>]
Arguments (in order):
	1. Telegram chat name
	2. Telegram id
"""
}

if [ $# -ne 2 ];then
	usage
	exit 1
fi

name="$1"
id="$2"

id=$(echo "${id}" | sed 's/-//g')

template="""
  # ${name}
  redis_${id}:
    container_name: redis_${id}
    image: redis:5.0-rc
  bot_${id}:
    build:
      context: ../
      args:
        PROJECT_PATH: github.com/novitoll/novitoll_daemon_bot
        TARGET: bot
    image: vahter-bot:0.0.9
    container_name: bot_${id}
    environment:
      - REDIS_HOST=redis_${id}
      - REDIS_PORT=6379
      - APP_LANG=rus
    env_file:
      - ./.env
    links:
      - redis_${id}
    depends_on:
      - redis_${id}
    volumes:
      - "../:/opt/src/github.com/novitoll/novitoll_daemon_bot"  
"""

echo "${template}" >> ./docker-compose.yml
cat ./docker-compose.yml