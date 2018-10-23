#!/bin/bash

set -e

token=$1
method=$2

function usage() {
	echo """./webhook.sh [<arguments>]
Arguments (in order):
	1. Telegram bot token
	2. Telegram bot method (setWebhook, getWebhookInfo etc."""
}

if [ $# -ne 2 ];then
	usage
	exit 1
fi

curl -X GET "https://api.telegram.org/bot${token}/${method}"