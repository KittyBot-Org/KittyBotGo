FROM golang:1.18-alpine AS build

ARG VERSION=dev

WORKDIR /tmp/app

COPY . .

RUN go mod download && \
    go mod verify && \
    go build -ldflags="-X 'main.version=${VERSION}'" -o dbot cmd/dbot/main.go

FROM alpine

WORKDIR /app

COPY --from=build /tmp/app/bot /app/bot

ENTRYPOINT ["./bot"]
