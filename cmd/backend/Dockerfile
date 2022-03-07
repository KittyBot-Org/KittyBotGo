FROM golang:1.17-alpine AS build

ARG VERSION=dev

WORKDIR /tmp/app

COPY . .

RUN go mod download && \
    go mod verify && \
    go build -ldflags="-X 'main.version=${VERSION}'" -o backend cmd/backend/main.go

FROM alpine

WORKDIR /app

COPY --from=build /tmp/app/backend /app/backend

ENTRYPOINT ["./backend"]
