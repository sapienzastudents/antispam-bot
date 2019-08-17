
.PHONY: docker clean dev push

antispam-telegram-bot:
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.APP_VERSION=${shell git describe --tags --dirty}" -a -installsuffix cgo -o antispam-telegram-bot .
	strip antispam-telegram-bot
	upx -9 antispam-telegram-bot

dev:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o antispam-telegram-bot .
	REDIS_URL=redis://127.0.0.1:6379/0 BOT_TOKEN="${BOT_TOKEN}" ./antispam-telegram-bot

docker:
	docker build -t antispam-telegram-bot:latest --build-arg APPVERSION="${shell git describe --tags --dirty}" \
		-f Dockerfile .

push:
	docker tag antispam-telegram-bot:latest enrico204/antispam-telegram-bot:latest
	docker push enrico204/antispam-telegram-bot:latest
	docker rmi enrico204/antispam-telegram-bot:latest

clean:
	rm -f antispam-telegram-bot