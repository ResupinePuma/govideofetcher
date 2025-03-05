package ytdl

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"videofetcher/internal/counting_reader"
	cr "videofetcher/internal/counting_reader"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/options"

	v "videofetcher/internal/downloader/video"
)

var (
	YTDlPath = "yt-dlp"
	Logging  Logger
)

var Extractors map[string][]string

type YtDl struct {
	SizeLimit int64
	Timeout   int

	Format     string
	FFmpeg     string
	Executable string
	Headers    http.Header
}

func NewParser(sizelim int64, opts *options.YTDLOptions) *YtDl {
	yt := YtDl{
		SizeLimit: sizelim,
		Headers:   http.Header{},
		FFmpeg:    opts.FFmpeg,
	}

	yt.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0")
	yt.Headers.Add("Accrpt-Language", "en-US,en;q=0.5") // Set default language to English for the user

	if opts != nil {
		yt.Format = opts.Format
		yt.Executable = opts.Executable
		yt.Headers = opts.Headers
	} else {
		yt.Executable = "yt-dlp"
		yt.Format = "18/17,bestvideo[height<=720][ext=mp4]+worstaudio,(mp4)[ext=mp4][vcodec^=h26],worst[width>=480][ext=mp4],worst[ext=mp4]"
	}
	return &yt
}

func randSize() int {
	min := 1 * 1 << 20
	max := 3 * 1 << 20
	return rand.Intn(max-min) + min
}

func (yt *YtDl) Download(ctx *dcontext.Context) (err error) {

	defer close(ctx.Results())

	u := ctx.GetUrl()

	tempPath, tempErr := os.MkdirTemp("", "ydls")
	if tempErr != nil {
		return tempErr
	}

	//dr := dresult.NewDownloaderResult(ctx)

	cmd := exec.CommandContext(
		ctx,
		YTDlPath,
		"--no-call-home",
		"--no-cache-dir",
		"--ignore-errors",
		"--max-filesize", "50M",
		"--newline",
		"--restrict-filenames",
		"-f", yt.Format, // скачиваем лучшее качество в формате mp4
	)

	if yt.Headers != nil {
		for k, v := range yt.Headers {
			line := fmt.Sprintf("%s: %s", k, strings.Join(v, "; "))
			cmd.Args = append(cmd.Args, "--add-header", line)
		}
	}
	if yt.FFmpeg != "" {
		cmd.Args = append(cmd.Args, "--exec", fmt.Sprintf("%s -i {} %s %s", "ffmpeg", yt.FFmpeg, "res.mp4"))
		cmd.Args = append(cmd.Args, u.String())
	} else {
		cmd.Args = append(cmd.Args, u.String())
		cmd.Args = append(cmd.Args, "-o", "res.mp4")
	}

	// var w io.WriteCloser
	// var r io.ReadCloser
	// //var errbuf bytes.Buffer
	// r, w = io.Pipe()

	cmd.Dir = tempPath
	cmd.Stderr = log.Writer()
	cmd.Stdout = log.Writer()
	// r, err = cmd.StdoutPipe()
	// if err != nil {
	// 	return
	// }

	//go func() {
	// defer os.Remove(tempPath)
	// defer w.Close()

	ctx.Notifier().UpdTextNotify("‍⏬ downloading video")
	go ctx.Notifier().StartTicker(ctx)

	Logging.Errorf("%v", cmd.Args)
	if err = cmd.Start(); err != nil {
		return
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	de, err := os.ReadDir(tempPath)
	if err != nil {
		return err
	}
	if len(de) == 0 {
		return fmt.Errorf("video not found")
	}

	vid := de[0].Name()

	rdr, err := os.Open(filepath.Join(tempPath, vid))
	if err != nil {
		return fmt.Errorf("err open result: %v", err)
	}

	cropts := cr.CountingReaderOpts{
		ByteLimit: yt.SizeLimit,
		FileSize:  float64(randSize()),
	}

	ctx.AddResult([]v.Video{{
		Reader: counting_reader.NewCountingReader(rdr, &cropts),
		Title:  u.String(),
		URL:    u.String(),
		Dir:    filepath.Join(tempPath, vid),
	}})

	return nil
}

func (yt *YtDl) Close() error {
	return nil
}
