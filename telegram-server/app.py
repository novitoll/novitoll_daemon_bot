from __future__ import print_function

import sys
import requests
import redis
import re

from flask import Flask, request

app = Flask(__name__)

URL = "https://api.telegram.org/bot{}/{}"
TCP_PORT = 5001
TELEGRAM_BOT_TOKEN = sys.argv[1]


# REDIS
REDIS_HOSTNAME = "redis"
REDIS_PORT = 6379
EXPIRATION = 14 * 24 * 3600  # 2 weeks


url_rgxp = re.compile('https?://(?:[-\w.]|(?:%[\da-fA-F]{2}))+')


class RedisClient(object):
    def __init__(self):
        self.conn = redis.StrictRedis(host=REDIS_HOSTNAME, port=REDIS_PORT)

    def set(self, k, v, ex=EXPIRATION):
        self.conn.set(k, v, ex)

    def get(self, k):
        return self.conn.get(k)


def notify_admin(exc):
    print(exc)
    pass


@app.route("/blogpost", methods=['GET', 'POST'])
def blogpost():
    method = "sendMessage"

    # 1. Bot incoming json pre-processing
    try:
        data = request.json
        msg = data['message']
        username = msg['chat']['username'] if 'username' in msg['chat'] else 'whothefookizit?'

        if 'entities' in msg and any(filter(lambda x: 'type' in x and x['type'] == 'url', msg['entities'])):
            if len(msg['text'].split(' ')) > 1:
                urls = url_rgxp.findall(msg['text'])
            else:
                urls = [msg['text']]
        else:
            return "response"

        _redis = RedisClient()
    except KeyError as exc:
        notify_admin(exc)
        return "bad response"

    # 2. check in Redis iteratively for URL duplication
    duplicates = filter(lambda x: not _redis.get(x), urls)

    if duplicates:
        reply = {
            'chat_id': msg['chat']['id'],
            'text': "@%s posted this URL %s as well. Please dont flood" % (username, msg['date']),
            'reply_to_message_id': msg['message_id'],
            'reply_markup': {
                'force_reply': True
            }
        }
        # sync
        response = requests.post(URL.format(TELEGRAM_BOT_TOKEN, method), json=reply)
    else:
        _redis.set(msg, {username: msg['date']})

    return "response"


if __name__ == "__main__":
    init()
    app.run(port=TCP_PORT)
