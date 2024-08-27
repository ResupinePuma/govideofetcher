package ytdl

import (
	"net/http"
	"net/url"
	cr "videofetcher/internal/counting_reader"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/derrors"
	"videofetcher/internal/downloader/options"

	v "videofetcher/internal/downloader/video"

	"github.com/ResupinePuma/goutubedl"
)

var (
	Logger iLogger
)

var Extractors map[string][]string

type YtDl struct {
	SizeLimit int64
	Timeout   int

	Format     string
	Executable string
	Headers    http.Header

	downloadResult *goutubedl.DownloadResult
}

func NewParser(sizelim int64, opts *options.YTDLOptions) *YtDl {
	yt := YtDl{
		SizeLimit: sizelim,
		Headers:   http.Header{},
	}

	yt.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0")
	yt.Headers.Add("Accrpt-Language", "en-US,en;q=0.5")

	if opts != nil {
		yt.Format = opts.Format
		yt.Executable = opts.Executable
	} else {
		yt.Executable = "yt-dlp"
		yt.Format = "18/17,bestvideo[height<=720][ext=mp4]+worstaudio,(mp4)[ext=mp4][vcodec^=h26],worst[width>=480][ext=mp4],worst[ext=mp4]"
	}
	return &yt
}

func (yt *YtDl) Download(ctx dcontext.Context, u *url.URL) (res []v.Video, err error) {
	//parts := strings.Split(u.Host, ".")
	// extr := []string{"generic"}
	// for i := 1; i < len(parts)+1; i++ {
	// 	e := strings.Join(parts[:i], ".")
	// 	exx, found := Extractors[e]
	// 	if found {
	// 		extr = exx
	// 		break
	// 	}
	// }

	goutubedl.Path = "yt-dlp"
	result, err := goutubedl.New(&ctx, u.String(), goutubedl.Options{
		Type:        goutubedl.TypeSingle,
		DebugLog:    Logger,
		HttpHeaders: yt.Headers,
		//Extractors:  extr,
	})
	if err != nil {
		return
	}

	if result.Info.FilesizeApprox >= float64(yt.SizeLimit) || result.Info.Format.Filesize >= float64(yt.SizeLimit) {
		err = derrors.ErrSizeLimitReached
		return
	}

	ctx.Notifier().UpdTextNotify("‍⏬ downloading video")
	yt.downloadResult, err = result.Download(&ctx, yt.Format)
	cropts := cr.CountingReaderOpts{
		ByteLimit: yt.SizeLimit,
		FileSize:  float64(yt.downloadResult.Size()),
		Notifier:  ctx.Notifier(),
	}

	res = append(res,
		*v.NewVideo(result.Info.Title, u.String(), cr.NewCountingReader(yt.downloadResult, &cropts)),
	)

	return
}

func (yt *YtDl) Close() error {
	if yt.downloadResult != nil {
		return yt.downloadResult.Close()
	}
	return nil
}
