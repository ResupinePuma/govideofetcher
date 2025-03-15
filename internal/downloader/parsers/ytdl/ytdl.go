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
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"videofetcher/internal/counting_reader"
	cr "videofetcher/internal/counting_reader"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/media"
	"videofetcher/internal/downloader/options"

	v "videofetcher/internal/downloader/media"
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

	cmd := exec.CommandContext(
		ctx,
		YTDlPath,
		"--no-call-home",
		"--no-cache-dir",

		"--proxy", yt.ProxyURL,
		"--newline",
		"--write-thumbnail", "--convert-thumbnails", "png",
		"--buffer-size", "10M",
		"-N", "8",
		"--restrict-filenames",
		"--write-info-json",
		"-f", yt.Format, // скачиваем лучшее качество в формате mp4
		"--max-filesize", "50M",
	)

	if yt.Headers != nil {
		for k, v := range yt.Headers {
			line := fmt.Sprintf("%s: %s", k, strings.Join(v, "; "))
			cmd.Args = append(cmd.Args, "--add-header", line)
		}
	}
	ext := ""
	switch yt.mode {
	case modeAudio:
		ext = ".mp3"
	case modeVideo:
		ext = ".mp4"
	}

	filename := randomFileName(8)

	if yt.FFmpeg != "" {
		cmd.Args = append(cmd.Args, "-o", filename+ext)
		cmd.Args = append(cmd.Args, "--exec", fmt.Sprintf("%s -i {} %s %s", "ffmpeg", yt.FFmpeg, "conv_"+filename+ext))
		cmd.Args = append(cmd.Args, u.String())
	} else {
		cmd.Args = append(cmd.Args, u.String())
		cmd.Args = append(cmd.Args, "-o", filename+ext)
	}

	cmd.Dir = tempPath
	cmd.Stderr = log.Writer()
	cmd.Stdout = log.Writer()

	ctx.Notifier().UpdTextNotify("‍⏬ downloading media")
	go ctx.Notifier().StartTicker(ctx)

	Logging.Debugf("%v", cmd.Args)
	if err = cmd.Start(); err != nil {
		return
	}

	var info Info
	var errch chan error = make(chan error)
	var wg sync.WaitGroup

	var size float64

	wg.Add(1)
	go func() {
		defer wg.Done()
		// will be doanloadad before main video
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second)
				// open info file
				infof, err := findFile(filepath.Join(tempPath, "*.info.json"))
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						continue
					}
					errch <- err
					return
				}
				i, err := os.Open(infof)
				if err != nil {
					errch <- fmt.Errorf("err open info: %v", err)
					return
				}
				err = json.NewDecoder(i).Decode(&info)
				if err != nil {
					errch <- fmt.Errorf("err decode info: %v", err)
					return
				}
				i.Close()

				size = max(info.Filesize, info.FilesizeApprox)
				if size == 0 {
					size = float64(randSize())
				}
				if size > float64(yt.SizeLimit) {
					errch <- fmt.Errorf("size limit reached")
					return
				}

				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err = cmd.Wait()
		if err != nil {
			errch <- err
		}

	}()

	go func() {
		wg.Wait()
		close(errch)
	}()

	for e := range errch {
		if e != nil {
			return e
		}
	}

	var thumbnail io.Reader
	if info.Thumbnail != "" {
		tbf, e := findFile(filepath.Join(tempPath, "*.png"))
		if e == nil {
			b, e := os.ReadFile(tbf)
			if e == nil {
				thumbnail = bytes.NewReader(b)
			}
		}
	}

	// open media file
	ptrn := "*" + ext
	if yt.FFmpeg != "" {
		ptrn = "conv_" + ptrn
	}

	vidf, err := findFile(filepath.Join(tempPath, ptrn))
	if err != nil {
		return err
	}
	rdr, err := os.Open(vidf)
	if err != nil {
		return fmt.Errorf("err open result: %v", err)
	}

	cropts := cr.CountingReaderOpts{
		ByteLimit: yt.SizeLimit,
		FileSize:  float64(size),
	}

	var m []media.Media
	switch yt.mode {
	case modeVideo:
		m = append(m, &v.Video{
			Reader:    counting_reader.NewCountingReader(rdr, &cropts),
			Title:     info.Title,
			Thumbnail: thumbnail,
			URL:       u.String(),
			Dir:       tempPath,
			Duration:  info.Duration,
			Filename:  info.Title + ".mp4",
		})
	case modeAudio:
		m = append(m, &v.Audio{
			Reader:    counting_reader.NewCountingReader(rdr, &cropts),
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
