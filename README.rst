@novitoll_daemon_bot - Telegram bot
=====

.. contents::

Features
-------

Bot features can be enabled/disabled via `config/features.json` and its ad-hoc struct `config/features.go`.

* standalone GoLang binary
* duplicate hyperlinks detection within the certain amount of time (2 weeks)
* ad detection (TBD)
* nude, pornography detection in image (TBD)
* newcomer questionnaire in bot's IM to prevent newcomers' shadow mode (TBD)

Make commands
-------
* `make configure` -- configure `dep` GoLang package
* `make install` -- dep installs
* `make build` -- compile Go src to the "$PWD/bot" binary
* `make run` -- compile and run
* `make test` -- run unit tests
* `make debug` -- compile and run `delve` debugger

License
------
GNU GPL 2.0
