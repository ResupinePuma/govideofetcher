package downloader

import (
	"context"
	"errors"
	"io"

	"github.com/ResupinePuma/goutubedl"
)

type YTdlLog struct {
	AbstractLogger
}

func (l *YTdlLog) Print(v ...interface{}) {
	l.Debug("%v", v...)
}

type YTdl struct {
	SizeLimit int `yaml:"-"`
	Timeout   int `yaml:"-"`

	Format     string "yaml:\"format\""
	Executable string "yaml:\"-\""

	downloadResult *goutubedl.DownloadResult
	log            YTdlLog
	ntf            AbstractNotifier
}

func (yt *YTdl) Init(logger AbstractLogger, notifier AbstractNotifier, opts *DownloaderOpts) error {
	yt.log = YTdlLog{
		AbstractLogger: logger,
	}
	yt.Timeout = opts.Timeout
	yt.SizeLimit = opts.SizeLimit
	yt.ntf = notifier
	return nil
}

func (yt *YTdl) Download(ctx context.Context, url string) (title string, rdr io.ReadCloser, err error) {
	yt.ntf.Message("‍⏬ downloading video")
	goutubedl.Path = "yt-dlp"
	result, err := goutubedl.New(ctx, url, goutubedl.Options{
		Type:     goutubedl.TypeSingle,
		DebugLog: &yt.log})
	if err != nil {
		return
	}

	if result.Info.FilesizeApprox >= float64(yt.SizeLimit) {
		err = errors.New("size limit reached")
		return
	}

	yt.downloadResult, err = result.Download(ctx, yt.Format)
	cropts := CountingReaderOpts{
		ByteLimit: yt.SizeLimit,
		FileSize:  float64(yt.downloadResult.Size()),
		Notifier:  yt.ntf,
	}

	return result.Info.Title, NewCountingReader(yt.downloadResult, &cropts), err
}

func (yt *YTdl) Close() error {
	if yt.downloadResult != nil {
		return yt.downloadResult.Close()
	}
	return nil
}
