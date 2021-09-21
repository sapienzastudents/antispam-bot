SHELL := /bin/bash

VERSION := $(shell git fetch --unshallow >/dev/null 2>&1; git describe --all --long --dirty 2>/dev/null)
ifeq (${VERSION},)
VERSION := no-git-version
endif

BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")


.PHONY: all
all: docker

.PHONY: docker
docker:
	docker build \
		-t antispam-telegram-bot:latest \
		--build-arg APP_VERSION="${VERSION}" \
		--build-arg BUILD_DATE=${BUILD_DATE} \
		.

.PHONY: push
push:
	docker tag antispam-telegram-bot:latest enrico204/antispam-telegram-bot:latest
	docker push enrico204/antispam-telegram-bot:latest

.PHONY: up-deps
up-deps:
	docker-compose \
		-f demo/docker-compose.yml \
		up

.PHONY: stop
stop:
	docker-compose \
		-f demo/docker-compose.yml \
		stop

.PHONY: down
down:
	docker-compose \
		-f demo/docker-compose.yml \
		down

.PHONY: test
test:
	go test ./... -mod=mod
	go vet ./...
	gosec -quiet ./...
	staticcheck -tests=false ./...
	ineffassign ./...
	errcheck ./...
	go list -u -m -json all | go-mod-outdated -update -direct

.PHONY: deploy
deploy:
	kubectl -n default set image deploy/antispam-tbot antispam-tbot=enrico204/antispam-telegram-bot@$(shell skopeo inspect docker://enrico204/antispam-telegram-bot:latest | jq -r ".Digest")

dev:
	CGO_ENABLED=0 GOOS=linux REDIS_URL=redis://127.0.0.1:6379/0 BOT_TOKEN="${BOT_TOKEN}" go run ./cmd/telegram/
