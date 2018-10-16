configure:
	go get -u github.com/golang/dep/cmd/dep
	[ -L  /usr/local/bin/dep ] || ln -s ${GOPATH}/bin/dep /usr/local/bin/dep
	dep ensure
	go get -u mvdan.cc/xurls github.com/go-redis/redis
	go get -u github.com/stretchr/testify/assert
	go get -u github.com/justincampbell/timeago

build:
	go build -o bot.bin cmd/novitoll_daemon_bot/main.go

run:
	@make build
	./bot.bin

docker-compose-dev:
	docker-compose -f deployments/docker-compose-dev.yml up

docker-compose:
	docker-compose -f deployments/docker-compose.yml up -d

debug:
	dlv debug --output bot.bin cmd/novitoll_daemon_bot/main.go

test:
	go test -v -cover -coverprofile=coverage.out github.com/novitoll/novitoll_daemon_bot/internal/vahter/bot
	go tool cover -html=coverage.out -o coverage.html 
