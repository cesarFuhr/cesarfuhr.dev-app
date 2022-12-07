# syntax=docker/dockerfile:1

FROM golang:1.16-alpine AS builder

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY main.go ./
COPY public ./public

RUN CGO_ENABLED=0 go build -o /build/server

FROM scratch AS runner

WORKDIR /runner
COPY --from=builder /build/server /runner/server

EXPOSE 80

CMD [ "/runner/server", "-HTTP_PORT=80"]
