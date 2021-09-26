FROM docker.io/library/golang:1.16 AS builder

#### Set Go environment
# Disable CGO to create a self-contained executable
# Do not enable unless it's strictly necessary
ENV CGO_ENABLED 0
# Set Linux as target
ENV GOOS linux

### Prepare base image
RUN apt-get update && apt-get install -y upx-ucl zip ca-certificates tzdata
RUN useradd --home /app/ -M appuser

WORKDIR /src/

### Copy Go modules files and cache dependencies
# If dependencies do not changes, these two lines are cached (speed up the build)
COPY go.* ./
RUN go mod download

### Copy Go code
COPY cmd cmd
COPY service service

### Set some build variables
ARG APP_VERSION
ARG BUILD_DATE

RUN go generate ./...

### Build executables, strip debug symbols and compress with UPX
WORKDIR /src/cmd/
RUN mkdir /app/
RUN /bin/bash -c "for ex in \$(ls); do pushd \$ex; go build -mod=readonly -ldflags \"-extldflags \\\"-static\\\" -X main.AppVersion=${APP_VERSION} -X main.BuildDate=${BUILD_DATE}\" -a -installsuffix cgo -o /app/\$ex .; popd; done"
RUN cd /app/ && strip * && upx -9 *

# We need to start from Debian as we need the SSH agent
FROM debian:buster

EXPOSE 3000

RUN apt-get update && \
    apt-get install -y ca-certificates tzdata ssh-client && \
    rm -rf /var/cache/apt/* && \
    useradd -d /app/ -M appuser

WORKDIR /app/
COPY docker-cmd.sh .
COPY --from=builder /app/* ./

USER appuser
CMD ["/app/docker-cmd.sh", "/app/telegram"]
