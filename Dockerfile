# syntax=docker/dockerfile:1

FROM golang:1.16-alpine

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY main.go ./

RUN mkdir /app
RUN go build -o /app/main

COPY public /app/public

RUN  mkdir /app/certs
COPY certs/cesarfuhr.crt /app/certs
COPY certs/cesarfuhr.key /app/certs
COPY certs/wildcesarfuhr.crt /app/certs
COPY certs/wildcesarfuhr.key /app/certs

WORKDIR /app

EXPOSE 443

CMD [ "/app/main", "-PORT=443" ]
