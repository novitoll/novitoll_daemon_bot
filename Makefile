configure:
	go get -u github.com/golang/dep/cmd/dep
	sudo ln -s ${GOPATH}/bin/dep /usr/local/bin/dep

install:
	dep ensure
	go get -u mvdan.cc/xurls github.com/go-redis/redis

build:
	go build cmd/novitoll_daemon_bot/main.go
