FROM golang:1.19-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o bot cmd/bot/bot.go

FROM alpine

COPY --from=build /build/bot /bin/bot

ENTRYPOINT ["/bin/bot"]

CMD ["-config", "/var/lib/bot/config.json"]
