FROM golang:1.20-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o backend cmd/backend/backend.go

FROM alpine

COPY --from=build /build/backend /bin/backend

ENTRYPOINT ["/bin/gateway"]

CMD ["-config", "/var/lib/backend/config.json"]
