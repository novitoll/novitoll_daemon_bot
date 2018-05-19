Telegram bot [<future bot name>]
=====

<put the description later>

.. contents::

Iteration 0
-------------

Goals:

* standalone GoLang binary
* duplicate hyperlinks detection within the certain amount of time


* setup a GoLang binary that does ``getUpdates``, e.g. via LongPolling gets the list of recent updates of the channel messages
* filter these messages and get only HTTP(S) hyperlinks
* save those links somewhere in the file for now (Iteration 1: implement a DB ~ MongoDB / Cassandra)
* return the response of duplicated hyperlinks

Backlog
-------------

* setup a web-server with GoLang (use external lib for Telegram bot API)
* setup a CloudFormation for the web-server for 2 EC2 with ELB that are bound with WebHooks
