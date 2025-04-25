FROM --platform=$BUILDPLATFORM golang:1.24-alpine3.20 AS builder

ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETOS

WORKDIR /govf

RUN apk add --no-cache curl ca-certificates build-base sqlite

COPY ./internal /govf/internal
COPY ./cmd /govf/cmd
COPY ./go.* /govf

RUN go mod download

RUN case "${TARGETPLATFORM}" in \
        "linux/amd64")  GOARCH=amd64 ;; \
        "linux/arm64")  GOARCH=arm64 ;; \
        "linux/arm/v7") GOARCH=arm  ;; \
        *) echo "Unsupported platform: ${TARGETPLATFORM}" && exit 1 ;; \
    esac && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$GOARCH go build \
        -o build/videofetcher \
        -ldflags='-s -w -extldflags "-static"' \
        cmd/videofetcher/main.go

FROM --platform=$BUILDPLATFORM alpine:3.20
WORKDIR /govf
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /govf/build/videofetcher /govf/videofetcher

ENTRYPOINT ["./videofetcher"]