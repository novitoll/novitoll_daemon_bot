import re
import requests
import pendulum

from flask import request

from vahter.bot.base import Bot
from vahter.wsgi import app, LOGGER, TELEGRAM_BOT_API_URL, TELEGRAM_BOT_TOKEN, TIMEZONE
from vahter.redis_client.base import RedisClient

url_rgxp = re.compile('https?://(?:[-\w.]|(?:%[\da-fA-F]{2}))+')


def notify_admin(exc, msg):
    LOGGER.error(exc)
    pass


@app.route("/vahter", methods=['POST'])
def vahter():
    method = "sendMessage"
    msg = None

    # 1. Bot incoming json pre-processing
    try:
        data = request.json
        msg = data['message']
        username = msg['chat']['username'] if 'username' in msg['chat'] else 'whothefookizit?'
        msg_dt = pendulum.from_timestamp(msg['date'], tz=TIMEZONE)

        LOGGER.debug("Message from %s" % username)

        if 'entities' in msg and any(filter(lambda x: 'type' in x and x['type'] == 'url', msg['entities'])):
            if len(msg['text'].split(' ')) > 1:
                urls = url_rgxp.findall(msg['text'])
            else:
                urls = [msg['text']]
        else:
            return "not url"

        # 2. connect to Redis
        _redis = RedisClient()
    except KeyError as exc:
        notify_admin(exc, msg)
        return "key error exception"

    # 3. check in Redis iteratively for URL duplication
    duplicates = filter(lambda x: not _redis.get(x), urls)

    if duplicates:
        # 4. Send the duplicate message
        reply = {
            'chat_id': msg['chat']['id'],
            'text': "@%s posted this URL %s as well. Please dont flood" % (username, msg_dt.subtract(days=1).diff_for_humans()),
            'reply_to_message_id': msg['message_id'],
            'reply_markup': {
                'force_reply': True
            }
        }

        # NB!: there is no need to use aiohttp for async HTTP I/O, because we can scale horizontaly with docker
        response = requests.post(TELEGRAM_BOT_API_URL.format(TELEGRAM_BOT_TOKEN, method), json=reply)
    else:
        # 4. Persist the message
        LOGGER.info("New url is set from %s - %s" % (username, msg['text']))
        _redis.set(msg['text'], {username: msg['date']})

    return "end of /vahter"

