@novitoll_daemon_bot - Telegram bot
=====

.. contents::

Requirements
-------

```
go version 
go1.11.1 linux/amd64
```

Features
-------

Bot features can be enabled/disabled via `config/features.json` and its ad-hoc struct `config/features.go`.

* duplicate hyperlinks detection within the certain amount of time
	* kindly reply with a notification
* newcomer questionnaire in bot's IM to prevent newcomers' shadow mode and post-action
	* greet a newcomer and kindly ask for the feedback upon the group joining in order to authenticate, otherwise user will be kicked for the certain time.
* stickers detection and post-action
	* kindly reply with a notification
	* can be configured to auto-delete the message with the sticker
* ad detection (TBD)
* nude, pornography detection in image (TBD)
* batch scanning of users' avatars, and posting images' for steganography analysis (+ VirusTotal?) (TBD)

Make commands
-------
* `make configure` -- configure `dep` GoLang package and install deps.
* `make build` -- compile Go src to the "$PWD/bot" binary.
* `make run` -- compile and run a standalone Go binary.
* `make docker-compose-local` -- For local development. Runs docker-compose that brings up 1 redis & 1 vahter-bot containers & 1 telegram-mock image.
* `make docker-compose` -- For the prod. Runs docker-compose that brings up 1 redis & 1 vahter-bot containers. `TELEGRAM_TOKEN` should be set as ENV var manually or in `.env` file. Containers run in deattached mode.
* `make docker-compose-green` -- For the green deployment running along with `make docker-compose` but on different TCP 8081 host port. Runs docker-compose that brings up 1 redis-dev & 1 vahter-bot-dev containers. `TELEGRAM_TOKEN` should be set as ENV var manually or in `.env` file.
* `make docker-compose-stop` -- Stops the containers run via `make docker-compose`.
* `make test` -- run unit tests.
* `make debug` -- compile and run `delve` debugger.

Flow
-------

.. image:: docs/flow.jpg

License
-------
GNU GPL 2.0