FROM golang:1.15 as builder

# Prerequisites for builds and scratch
RUN apt-get update && apt-get install -y upx-ucl zip ca-certificates tzdata

WORKDIR /src/

ARG APPVERSION

# Copy code and get dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-extldflags \"-static\" -X main.AppVersion=${APPVERSION}" -a -installsuffix cgo -o antispam-telegram-bot . && \
	strip antispam-telegram-bot && \
	upx -9 antispam-telegram-bot


# From empty container
FROM debian:buster

EXPOSE 3000

WORKDIR /app/

RUN apt-get update && \
    apt-get install -y ca-certificates tzdata ssh-client && \
    rm -rf /var/cache/apt/* && \
    useradd -d /app/ -M appuser

WORKDIR /app/
COPY docker-cmd.sh .
COPY --from=builder /src/antispam-telegram-bot .

USER appuser
CMD ["/app/docker-cmd.sh", "/app/antispam-telegram-bot"]
