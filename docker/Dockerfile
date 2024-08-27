FROM golang:1.22 as builder
ARG TARGETARCH

WORKDIR /govf
COPY . /govf
RUN apt-get update && apt-get install -y curl ca-certificates
RUN curl -L -o /govf/ffmpeg https://github.com/eugeneware/ffmpeg-static/releases/download/b5.0.1/ffmpeg-linux-${TARGETARCH}
RUN go build -o build/videofetcher --ldflags '-linkmode external -extldflags "-static"' cmd/videofetcher/main.go

FROM jauderho/yt-dlp:latest
WORKDIR /govf
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /govf/build/videofetcher /govf/videofetcher
COPY --from=builder /govf/ffmpeg /usr/local/bin/
RUN chmod +x /usr/local/bin/*

ENTRYPOINT ["./videofetcher"]
