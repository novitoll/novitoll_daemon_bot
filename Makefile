configure:
	go get -u github.com/golang/dep/cmd/dep
	sudo ln -s ${GOPATH}/bin/dep /usr/local/bin/dep

install:
	dep ensure
	go get -u mvdan.cc/xurls github.com/go-redis/redis
	go get -u github.com/stretchr/testify/assert

build:
	go build -o bot cmd/novitoll_daemon_bot/main.go

run:
	make build
	./bot

debug:
	dlv debug --output bot cmd/novitoll_daemon_bot/main.go

test:
	go test -v -cover -coverprofile=coverage.out github.com/novitoll/novitoll_daemon_bot/internal/vahter/bot
	go tool cover -html=coverage.out -o coverage.html 
