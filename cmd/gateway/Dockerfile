FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build -trimpath -o gateway ./cmd/gateway


FROM alpine

RUN apk add --no-cache ca-certificates

COPY --from=build /build/gateway /bin/gateway

ENTRYPOINT ["/bin/gateway"]

CMD ["-config", "/var/lib/gateway/config.toml"]
