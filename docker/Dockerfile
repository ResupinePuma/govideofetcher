FROM golang:1.22 as builder
WORKDIR /govf
COPY . /govf
RUN go build -o build/videofetcher --ldflags '-linkmode external -extldflags "-static"' cmd/videofetcher/main.go


FROM jauderho/yt-dlp:latest
WORKDIR /govf
COPY --from=builder /govf/build/videofetcher /govf/videofetcher
COPY --from=mwader/static-ffmpeg:latest /ffmpeg /usr/local/bin/

ENTRYPOINT ["./videofetcher"]
