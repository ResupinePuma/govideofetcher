FROM golang:1.22-alpine3.20 as builder
ARG TARGETARCH

WORKDIR /govf
RUN apk add --no-cache curl ca-certificates build-base sqlite
COPY . /govf
RUN if [ "$TARGETARCH" = "amd64" ]; then \
        curl -L -o /usr/local/bin/ffmpeg https://github.com/eugeneware/ffmpeg-static/releases/download/b6.0/ffmpeg-linux-x64 && \
        curl -L -o /usr/local/bin/ffprobe https://github.com/eugeneware/ffmpeg-static/releases/download/b6.0/ffprobe-linux-x64; \
    elif [ "$TARGETARCH" = "arm64" ]; then \
        curl -L -o /usr/local/bin/ffmpeg https://github.com/eugeneware/ffmpeg-static/releases/download/b6.0/ffmpeg-linux-arm64 && \
        curl -L -o /usr/local/bin/ffprobe https://github.com/eugeneware/ffmpeg-static/releases/download/b6.0/ffprobe-linux-arm64; \
    else \
        echo "Unsupported architecture: $TARGETARCH"; exit 1; \
    fi && \
    chmod +x /usr/local/bin/ffmpeg /usr/local/bin/ffprobe
RUN go build -o build/videofetcher --ldflags '-linkmode external -extldflags "-static"' cmd/videofetcher/main.go

FROM python:3-alpine3.20 as ytdlp-builder
WORKDIR /root
RUN apk add --no-cache wget binutils g++
RUN wget https://github.com/yt-dlp/yt-dlp/archive/refs/tags/$(wget -qO- https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest | grep '"tag_name"' | cut -d '"' -f 4).tar.gz && \
    tar -xzvf *.tar.gz && \
    cd yt-dlp-* && \
    python3 devscripts/install_deps.py --include pyinstaller && \
    python3 devscripts/make_lazy_extractors.py && \
    python3 -m bundle.pyinstaller && \
    cp dist/yt-dlp_linux* /root/yt-dlp

FROM python:3-alpine3.20
WORKDIR /govf
ARG TARGETARCH
RUN [ "$TARGETARCH" = "arm64" ] && apk add --no-cache v4l-utils libv4l || true
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /govf/build/videofetcher /govf/videofetcher
COPY --from=ytdlp-builder /root/yt-dlp /usr/local/bin/
COPY --from=builder /usr/local/bin/ffmpeg /usr/local/bin/
COPY --from=builder /usr/local/bin/ffprobe /usr/local/bin/
RUN chmod +x /usr/local/bin/*

ENTRYPOINT ["./videofetcher"]