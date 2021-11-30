# syntax=docker/dockerfile:1

FROM golang:1.16-alpine

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY main.go ./

RUN mkdir /app
RUN go build -o /app/main

COPY public /app/public
COPY full.crt /app
COPY priv.key /app

WORKDIR /app

EXPOSE 443

CMD [ "/app/main", "-PORT=443" ]
