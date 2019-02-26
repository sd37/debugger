FROM golang:alpine

RUN apk add git
RUN apk add --update binutils

COPY . /go/src/debugger

WORKDIR /go/src/debugger

RUN go get ./...


