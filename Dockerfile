FROM golang:1.18.2-alpine

RUN apk update && mkdir /go/src/app

# airのインストール
RUN go install github.com/cosmtrek/air@latest && \
  mv /go/bin/air /usr/local/bin/

# モジュールをあらかじめダウンロード
WORKDIR /go/src/app
COPY go.* .
RUN go mod download
