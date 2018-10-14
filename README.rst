@novitoll_daemon_bot - Telegram bot
=====

.. contents::

Requirements
-------

```
go version go1.11.1 linux/amd64
```

Features
-------

Bot features can be enabled/disabled via `config/features.json` and its ad-hoc struct `config/features.go`.

* duplicate hyperlinks detection within the certain amount of time (2 weeks)
* ad detection (TBD)
* nude, pornography detection in image (TBD)
* newcomer questionnaire in bot's IM to prevent newcomers' shadow mode (TBD)
* batch scanning of users' avatars, and posting images' for steganography analysis (+ VirusTotal?) (TBD)

Make commands
-------
* `make configure` -- configure `dep` GoLang package and install deps
* `make build` -- compile Go src to the "$PWD/bot" binary
* `make run` -- compile and run a standalone Go binary
* `make docker-compose` -- run docker-compose that brings up 1 redis & 1 vahter-bot containers
* `make test` -- run unit tests
* `make debug` -- compile and run `delve` debugger

TODO
-------
* Replace hostnames (currently localhost / mock containers are used)

Flow
-------

.. image:: docs/flow.jpg

License
-------
GNU GPL 2.0
