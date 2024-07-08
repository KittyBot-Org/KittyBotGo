FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ARG COMMIT

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build -ldflags="-X 'main.Version=$VERSION' -X 'main.Commit=$COMMIT'" -o bot github.com/KittyBot-Org/KittyBotGo/cmd/bot


FROM alpine

COPY --from=build /build/bot /bin/bot

ENTRYPOINT ["/bin/bot"]

CMD ["-config", "/var/lib/bot/config.yoml"]
