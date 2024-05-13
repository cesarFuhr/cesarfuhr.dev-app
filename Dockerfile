# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY cmd/blog/main.go ./
COPY cmd/blog/public ./public

RUN CGO_ENABLED=0 go build -o /build/server

FROM scratch AS runner

WORKDIR /runner
COPY --from=builder /build/server /runner/server

CMD [ "/runner/server", "-HTTP_PORT=8080"]
