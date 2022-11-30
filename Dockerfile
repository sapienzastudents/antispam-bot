FROM golang:1.19 as build

LABEL org.opencontainers.image.authors="inbox@emanuelepetriglia.com"

ENV GOARCH amd64

# Disable CGO to build a self-contained executable.
ENV CGO_ENABLED 0

WORKDIR /app

# Copy go.mod and go.sum before coping the code to cache the deps.
COPY go.* ./
RUN go mod download -x && go mod verify

# Two COPY instruction because if we give a directory as argument it copies the
# contents and not the directory itself.
COPY *.go ./
COPY service ./service
RUN go build -v -o antispam-telegram-bot .

FROM alpine:3.17

# Configuration file for the bot.
ENV ANTISPAM_PATH ./config.yml

COPY --from=build /app/antispam-telegram-bot /app/

ENTRYPOINT ["/app/antispam-telegram-bot"]

