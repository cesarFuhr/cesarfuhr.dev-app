# syntax=docker/dockerfile:1

FROM golang:1.21-alpine AS builder

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY main.go ./
COPY public ./public

RUN CGO_ENABLED=0 go build -o /build/server

FROM scratch AS runner

WORKDIR /runner
COPY --from=builder /build/server /runner/server

CMD [ "/runner/server", "-HTTP_PORT=8080"]
