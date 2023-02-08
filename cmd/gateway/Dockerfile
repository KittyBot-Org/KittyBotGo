FROM golang:1.19-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o gateway cmd/gateway/gateway.go

FROM alpine

COPY --from=build /build/gateway /bin/gateway

ENTRYPOINT ["/bin/gateway"]

CMD ["-config", "/var/lib/gateway/config.json"]
