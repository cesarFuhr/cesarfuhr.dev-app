# syntax=docker/dockerfile:1

FROM golang:1.16-alpine

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY main.go ./
COPY public ./public
COPY certs ./certs

RUN mkdir /app
RUN go build -o /app/main

COPY public /app/public

WORKDIR /app

EXPOSE 443

CMD [ "/app/main", "-PORT=443" ]
