package ytdl

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	YTDLPApi string
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
		YTDLPApi:  opts.APIAddr,
	}

	//yt.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:139.0) Gecko/20100101 Firefox/139.0")
	//yt.Headers.Add("Accept-Language", "en-US,en;q=0.5") // Set default language to English for the user

	if opts != nil {
		yt.Format = opts.Format
		yt.Executable = opts.Executable
		//yt.Headers = opts.Headers
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
		YTDLPApi:  opts.APIAddr,
	}

	//yt.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:139.0) Gecko/20100101 Firefox/139.0")
	//yt.Headers.Add("Accept-Language", "en-US,en;q=0.5") // Set default language to English for the user

	if opts != nil {
		yt.Executable = opts.Executable
		//yt.Headers = opts.Headers
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

// extByMode возвращает расширение файла по режиму
func extByMode(mode int) string {
	switch mode {
	case modeAudio:
		return ".mp3"
	case modeVideo:
		return ".mp4"
	default:
		return ""
	}
}

// postConvert сериализует args в JSON и шлёт POST-запрос
func postConvert(ctx context.Context, url string, args ytdlpapi.BodyArgs) (*http.Response, error) {
	body, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("json marshal args: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	return http.DefaultClient.Do(req)
}

// newMultipartReader создаёт multipart.Reader из ответа
func newMultipartReader(resp *http.Response) (*multipart.Reader, error) {
	ct := resp.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return nil, fmt.Errorf("parse media type: %w", err)
	}
	if !strings.HasPrefix(mediaType, "multipart/") {
		return nil, fmt.Errorf("expected multipart/*, got %s", mediaType)
	}
	boundary, ok := params["boundary"]
	if !ok {
		return nil, fmt.Errorf("missing boundary in content-type")
	}
	return multipart.NewReader(resp.Body, boundary), nil
}

// parseParts читает части multipart и возвращает инфо, thumb и media.Part
func parseParts(mr *multipart.Reader) (info Info, thumb io.Reader, mediaPart io.ReadCloser, err error) {
	for {
		part, errPart := mr.NextPart()
		if errors.Is(errPart, io.EOF) {
			break
		}
		if errPart != nil {
			return info, nil, nil, fmt.Errorf("reading multipart: %w", errPart)
		}

		switch part.FormName() {
		case "info":
			if err := json.NewDecoder(part).Decode(&info); err != nil {
				part.Close()
				return info, nil, nil, fmt.Errorf("decoding info: %w", err)
			}
			part.Close()

		case "thumb":
			buf, err := io.ReadAll(part)
			if err != nil {
				part.Close()
				continue
			}
			thumb = bytes.NewReader(buf)

		case "media":
			return info, thumb, part, nil
		default:
			part.Close()
		}
	}
	return info, thumb, nil, fmt.Errorf("no media part found")
}

// buildMediaSlice создаёт срез media.Media в зависимости от режима
func buildMediaSlice(mode int, reader io.ReadCloser, opts *cr.CountingReaderOpts, info Info, thumb io.Reader, dir, url string) []media.Media {
	crdr := counting_reader.NewCountingReader(reader, opts)
	switch mode {
	case modeVideo:
		return []media.Media{
			&media.Video{
				Reader:    crdr,
				Title:     info.Title,
				Thumbnail: thumb,
				URL:       url,
				Dir:       dir,
				Duration:  info.Duration,
				Filename:  info.Title + ".mp4",
			},
		}
	case modeAudio:
		filename := fmt.Sprintf("%s-%s.mp3", info.Artist, info.Title)
		return []media.Media{
			&media.Audio{
				Reader:    crdr,
				Title:     info.Title,
				Artist:    info.Artist,
				Thumbnail: thumb,
				URL:       url,
				Dir:       dir,
				Duration:  info.Duration,
				Filename:  filename,
			},
		}
	default:
		return nil
	}
}

func (yt *YtDl) Download(ctx *dcontext.Context) error {
	defer close(ctx.Results())

	// Подготовка
	u := ctx.GetUrl()
	tempPath, err := os.MkdirTemp("", "ydls")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}

	lang := ctx.GetLang()
	yt.Format = strings.ReplaceAll(yt.Format, "language={}", fmt.Sprintf("language^=%s", lang))

	ext := extByMode(yt.mode)

	// Формируем и отправляем запрос
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

	ctx.Notifier().UpdTextNotify("⏬ downloading media")
	go ctx.Notifier().StartTicker(ctx)

	resp, err := postConvert(ctx, yt.YTDLPApi, args)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	mr, err := newMultipartReader(resp)
	if err != nil {
		return err
	}

	info, thumb, mediaPart, err := parseParts(mr)
	if err != nil {
		return err
	}

	// Собираем результирующий media.Media
	cropts := cr.CountingReaderOpts{
		ByteLimit: yt.SizeLimit,
		FileSize:  float64(max(info.Filesize, info.FilesizeApprox)),
	}

	result := buildMediaSlice(yt.mode, mediaPart, &cropts, info, thumb, tempPath, u.String())

	ctx.AddResult(result)
	return nil
}

func (yt *YtDl) Close() error {
	return nil
}
