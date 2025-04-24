package ytdl

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"videofetcher/internal/counting_reader"
	cr "videofetcher/internal/counting_reader"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/media"
	"videofetcher/internal/downloader/options"
	ytdlpapi "videofetcher/internal/yt-dlp-api"
)

var (
	YTDlPath = "yt-dlp"
	Logging  Logger
)

const (
	modeVideo = 1
	modeAudio = 2

	mediaName = "res_media"
)

var Extractors map[string][]string

type YtDl struct {
	SizeLimit int64
	Timeout   int

	Format     string
	FFmpeg     string
	Executable string
	Headers    http.Header

	ProxyURL string
	mode     int
}

func randomFileName(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(err)
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b)
}

func NewParser(sizelim int64, opts *options.YTDLOptions) *YtDl {
	yt := YtDl{
		SizeLimit: sizelim,
		Headers:   http.Header{},
		mode:      modeVideo,
	}

	yt.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0")
	yt.Headers.Add("Accept-Language", "en-US,en;q=0.5") // Set default language to English for the user

	if opts != nil {
		yt.Format = opts.Format
		yt.Executable = opts.Executable
		yt.Headers = opts.Headers
		yt.FFmpeg = opts.FFmpeg
		yt.ProxyURL = opts.Proxies
	} else {
		yt.Executable = "yt-dlp"
		yt.Format = "18/17,bestvideo[height<=720][ext=mp4]+worstaudio,(mp4)[ext=mp4][vcodec^=h26],worst[width>=480][ext=mp4],worst[ext=mp4]"
	}
	return &yt
}

func NewParserAudio(sizelim int64, opts *options.YTDLOptions) *YtDl {
	yt := YtDl{
		SizeLimit: sizelim,
		Headers:   http.Header{},
		Format:    "bestaudio",
		mode:      modeAudio,
	}

	yt.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0")
	yt.Headers.Add("Accept-Language", "en-US,en;q=0.5") // Set default language to English for the user

	if opts != nil {
		yt.Executable = opts.Executable
		yt.Headers = opts.Headers
		yt.ProxyURL = opts.Proxies
	} else {
		yt.Executable = "yt-dlp"
	}
	return &yt
}

var ErrNotFound = fmt.Errorf("file not found")

func findFile(pattern string) (string, error) {
	// Поиск файлов по маске
	files, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	// Если файл найден, возвращаем первый
	if len(files) > 0 {
		return files[0], nil
	}
	return "", ErrNotFound
}

func randSize() int {
	min := 1 * 1 << 20
	max := 3 * 1 << 20
	r, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	r.Add(r, big.NewInt(int64(min)))
	return int(r.Int64())
}

func (yt *YtDl) Download(ctx *dcontext.Context) (err error) {

	defer close(ctx.Results())

	u := ctx.GetUrl()

	tempPath, tempErr := os.MkdirTemp("", "ydls")
	if tempErr != nil {
		return tempErr
	}

	lang := ctx.GetLang()
	yt.Format = strings.ReplaceAll(yt.Format, "language={}", fmt.Sprintf("language^=%s", lang))

	ext := ""
	switch yt.mode {
	case modeAudio:
		ext = ".mp3"
	case modeVideo:
		ext = ".mp4"
	}

	args := ytdlpapi.BodyArgs{
		URL:           u.String(),
		ProxyURL:      yt.ProxyURL,
		Format:        yt.Format,
		FFMpeg:        yt.FFmpeg,
		Headers:       yt.Headers,
		Extension:     ext,
		DownloadInfo:  true,
		DownloadThumb: true,
		BufferSize:    "10M",
		MaxFilesize:   "50M",
	}

	b, _ := json.Marshal(args)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:8080/convert", bytes.NewReader(b))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("BAD STATUS")
	}

	ctx.Notifier().UpdTextNotify("‍⏬ downloading media")
	go ctx.Notifier().StartTicker(ctx)

	contentType := resp.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Fatalf("ParseMediaType error: %v", err)
	}
	if mediaType != "multipart/form-data" && mediaType != "multipart/mixed" {
		log.Fatalf("Expected multipart/*, got %s", mediaType)
	}
	boundary, ok := params["boundary"]
	if !ok {
		log.Fatalf("No boundary in Content-Type")
	}

	// создаём Reader для разбора body
	mr := multipart.NewReader(resp.Body, boundary)

	var info Info
	var errch chan error = make(chan error)
	var thumbnail io.Reader
	var wg sync.WaitGroup
	var mreader io.ReadCloser

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			part, err := mr.NextPart()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				errch <- fmt.Errorf("err io error: %v", err)
			}
			switch part.FormName() {
			case "info":
				err = json.NewDecoder(part).Decode(&info)
				if err != nil {
					errch <- fmt.Errorf("err decode info: %v", err)
					continue
				}
				part.Close()

			case "thumb":
				b, err := io.ReadAll(part)
				if err != nil {
					errch <- fmt.Errorf("err decode thumb: %v", err)
					continue
				}
				part.Close()

				thumbnail = bytes.NewReader(b)
			case "media":
				mreader = part
				return
			}
		}
	}()

	wg.Wait()

	cropts := cr.CountingReaderOpts{
		ByteLimit: yt.SizeLimit,
		FileSize:  float64(max(info.Filesize, info.FilesizeApprox)),
	}

	var m []media.Media
	switch yt.mode {
	case modeVideo:
		m = append(m, &media.Video{
			Reader:    counting_reader.NewCountingReader(mreader, &cropts),
			Title:     info.Title,
			Thumbnail: thumbnail,
			URL:       u.String(),
			Dir:       tempPath,
			Duration:  info.Duration,
			Filename:  info.Title + ".mp4",
		})
	case modeAudio:
		m = append(m, &media.Audio{
			Reader:    counting_reader.NewCountingReader(mreader, &cropts),
			Title:     info.Title,
			URL:       u.String(),
			Thumbnail: thumbnail,
			Artist:    info.Artist,
			Dir:       tempPath,
			Duration:  info.Duration,
			Filename:  strings.Join([]string{info.Artist, info.Title}, "-") + ".mp3",
		})
	}
	ctx.AddResult(m)

	return nil
}

func (yt *YtDl) Close() error {
	return nil
}
