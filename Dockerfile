FROM golang:1.20 as builder
WORKDIR /govf
COPY ./ /govf

RUN go build --ldflags '-linkmode external -extldflags "-static"'

FROM jauderho/yt-dlp:latest
WORKDIR /govf
COPY --from=builder /govf/videofetcher /govf/videofetcher
COPY --from=mwader/static-ffmpeg:latest /ffmpeg /usr/local/bin/
COPY --from=mwader/static-ffmpeg:latest /ffprobe /usr/local/bin/

ENTRYPOINT ["./videofetcher"]
