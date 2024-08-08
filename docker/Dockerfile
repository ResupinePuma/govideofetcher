FROM golang:1.22 as builder
ARG TARGETARCH

WORKDIR /govf
COPY . /govf
RUN apt-get update && apt-get install -y curl ca-certificates
RUN curl -L -o /govf/ffmpeg https://github.com/eugeneware/ffmpeg-static/releases/download/b5.0.1/ffmpeg-linux-${TARGETARCH}
RUN go build -o build/videofetcher --ldflags '-linkmode external -extldflags "-static"' cmd/videofetcher/main.go

FROM ubuntu:20.04 as ytdlp-builder
WORKDIR /root
RUN apt-get update && apt-get install -y wget python3-dev python3-pip
RUN wget https://github.com/yt-dlp/yt-dlp/archive/refs/tags/2024.07.09.tar.gz && tar -xzvf 2024.07.09.tar.gz && \
    cd yt-dlp-2024.07.09 && \
    python3 devscripts/install_deps.py --include pyinstaller && \
    python3 devscripts/make_lazy_extractors.py && \
    python3 -m bundle.pyinstaller && \
    cp dist/yt-dlp_linux* /root/yt-dlp

FROM ubuntu:20.04
WORKDIR /govf
RUN apt update && apt install -y libpython3.8
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /govf/build/videofetcher /govf/videofetcher
COPY --from=ytdlp-builder /root/yt-dlp /usr/local/bin/
COPY --from=builder /govf/ffmpeg /usr/local/bin/
RUN chmod +x /usr/local/bin/*

ENTRYPOINT ["./videofetcher"]
