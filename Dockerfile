FROM golang:1.23.4

WORKDIR ${GOPATH}/mattermost-bot-vote/

# Устанавливаем зависимости для OpenSSL (для Debian-based образов)
RUN apt-get update && apt-get install -y libssl-dev pkg-config && rm -rf /var/lib/apt/lists/*

COPY . ${GOPATH}/mattermost-bot-vote/

RUN go build -o /build ./ \
    && go clean -cache -modcache

EXPOSE 8080

CMD ["/build"]