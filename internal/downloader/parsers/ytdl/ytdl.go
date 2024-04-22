package ytdl

import (
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

type YtDl struct {
	SizeLimit int64
	Timeout   int

	Format     string
	Executable string

	downloadResult *goutubedl.DownloadResult
}

func NewParser(sizelim int64, opts *options.YTDLOptions) *YtDl {
	yt := YtDl{
		SizeLimit: sizelim,
	}

	if opts != nil {
		yt.Format = opts.Format
		yt.Executable = opts.Executable
	} else {
		yt.Executable = "yt-dlp"
		yt.Format = "18/17,bestvideo[height<=720][ext=mp4]+worstaudio,(mp4)[ext=mp4][vcodec^=h26],worst[width>=480][ext=mp4],worst[ext=mp4]"
	}
	return &yt
}

func (yt *YtDl) Download(ctx dcontext.Context, u string) (res []v.Video, err error) {
	goutubedl.Path = "yt-dlp"
	result, err := goutubedl.New(&ctx, u, goutubedl.Options{
		Type:     goutubedl.TypeSingle,
		DebugLog: Logger})
	if err != nil {
		return
	}

	if result.Info.FilesizeApprox >= float64(yt.SizeLimit) {
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
		*v.NewVideo(result.Info.Title, u, cr.NewCountingReader(yt.downloadResult, &cropts)),
	)

	return
}

func (yt *YtDl) Close() error {
	if yt.downloadResult != nil {
		return yt.downloadResult.Close()
	}
	return nil
}
