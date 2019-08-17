FROM golang:1.12 as builder

# Prerequisites for builds and scratch
RUN apt-get update && apt-get install -y upx-ucl zip ca-certificates tzdata
WORKDIR /usr/share/zoneinfo
RUN zip -r -0 /zoneinfo.zip .
RUN adduser --home /app/ --no-create-home --disabled-password --quiet appuser

WORKDIR /src/

ARG APPVERSION

# Copy code and get dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-extldflags \"-static\" -X main.APP_VERSION=${APPVERSION}" -a -installsuffix cgo -o antispam-telegram-bot . && \
	strip antispam-telegram-bot && \
	upx -9 antispam-telegram-bot


# From empty container
FROM scratch

ENV ZONEINFO /zoneinfo.zip
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /zoneinfo.zip /
COPY --from=builder /etc/passwd /etc/passwd

WORKDIR /app/
COPY --from=builder /src/antispam-telegram-bot .

USER appuser
CMD ["/app/antispam-telegram-bot"]
