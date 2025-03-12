FROM golang:1.22-alpine3.20 as builder
ARG TARGETARCH

WORKDIR /govf
COPY . /govf
RUN apk update && apk upgrade -a && apk add --no-cache curl ca-certificates build-base sqlite
RUN if [ "$TARGETARCH" = "amd64" ]; then \
        curl -L -o /usr/local/bin/ffmpeg https://github.com/eugeneware/ffmpeg-static/releases/download/b6.0/ffmpeg-linux-x64; \
        curl -L -o /usr/local/bin/ffprobe https://github.com/eugeneware/ffmpeg-static/releases/download/b6.0/ffprobe-linux-x64; \
    elif [ "$TARGETARCH" = "arm64" ]; then \
        curl -L -o /usr/local/bin/ffmpeg https://github.com/eugeneware/ffmpeg-static/releases/download/b6.0/ffmpeg-linux-arm64; \
        curl -L -o /usr/local/bin/ffprobe https://github.com/eugeneware/ffmpeg-static/releases/download/b6.0/ffprobe-linux-arm64; \
    else \
        echo "Unsupported architecture: $TARGETARCH"; exit 1; \
    fi
RUN go build -o build/videofetcher --ldflags '-linkmode external -extldflags "-static"' cmd/videofetcher/main.go

FROM python:3-alpine3.20 as ytdlp-builder
WORKDIR /root
RUN apk update && apk upgrade -a && apk add wget binutils g++
RUN wget https://github.com/yt-dlp/yt-dlp/archive/refs/tags/2024.10.07.tar.gz && tar -xzvf 2024.10.07.tar.gz && \
    cd yt-dlp-2024.10.07 && \
    python3 devscripts/install_deps.py --include pyinstaller && \
    python3 devscripts/make_lazy_extractors.py && \
    python3 -m bundle.pyinstaller && \
    cp dist/yt-dlp_linux* /root/yt-dlp

FROM python:3-alpine3.20
WORKDIR /govf
RUN if [ "$TARGETARCH" = "arm64" ]; then \
    apk add --no-cache v4l-utils libv4l  && rm -rf /var/cache/apk/* \
    else \
      echo "$TARGETARCH"; \
    fi
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /govf/build/videofetcher /govf/videofetcher
COPY --from=ytdlp-builder /root/yt-dlp /usr/local/bin/
COPY --from=builder /usr/local/bin/ffmpeg /usr/local/bin/
COPY --from=builder /usr/local/bin/ffprobe /usr/local/bin/
RUN chmod +x /usr/local/bin/*

ENTRYPOINT ["./videofetcher"]
