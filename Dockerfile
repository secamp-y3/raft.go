FROM golang:1.20.6-bookworm

RUN apt update && mkdir /go/src/appp

RUN go install github.com/cosmtrek/air@latest
ENV PATH /go/bin:$PATH

WORKDIR /go/src/app
COPY go.* .
RUN go mod download
