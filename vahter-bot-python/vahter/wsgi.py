import logging
import os

from flask import Flask

app = Flask(__name__)

FORMAT = '%(asctime)-15s %(message)s'
logging.basicConfig(format=FORMAT, datefmt="%m/%d/%Y %I:%M:%S %p %Z")
LOGGER = logging.getLogger('tcpserver')

TIMEZONE = 'Asia/Almaty'

app._logger = LOGGER
app.logger_name = LOGGER.name

TELEGRAM_BOT_API_URL = "https://api.telegram.org/bot{}/{}"
TELEGRAM_BOT_TOKEN = os.environ['TELEGRAM_BOT_TOKEN']

LOGGER.info("Running for token bot %s..." % TELEGRAM_BOT_TOKEN[:4])
