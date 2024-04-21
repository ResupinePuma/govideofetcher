package downloader

import (
	"context"
	"regexp"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/parsers/tiktok"
	"videofetcher/internal/downloader/video"
	"videofetcher/internal/utils"
)

var (
	ParserTikTok  = "tt"
	ParserIG      = "ig"
	ParserDefault = "any"
)

type DownloaderOpts struct {
	SizeLimit int64 `json:"size_limit"`
	Timeout   int64 `json:"timeout"`
}

type Downloader struct {
	sizelimit int64
	timeout   int64

	notifier AbstractNotifier
	parsers  map[string]AbstractDownloader
}

func NewDownloader(opts *DownloaderOpts) *Downloader {
	d := &Downloader{
		sizelimit: 100 << 20, //100Mb
	}
	if opts != nil {
		d.sizelimit = opts.SizeLimit
		d.timeout = opts.Timeout
	}
	d.parsers = map[string]AbstractDownloader{
		ParserTikTok: tiktok.NewParser(d.sizelimit),
	}
	return d
}

func (d *Downloader) SetNotifier(n AbstractNotifier) {
	d.notifier = n
}

func (d *Downloader) Download(ctx context.Context, text string) (res []video.Video, err error) {

	url, label, err := utils.ExtractUrlAndText(text)
	if err != nil {
		return
	}

	var dwn AbstractDownloader
	switch url.Hostname() {
	case "tiktok.com", "vt.tiktok.com":
		dwn = d.parsers[ParserTikTok]
	}

	videos, err := dwn.Download(dcontext.NewContext(ctx, d.notifier), url.String())
	if err != nil {
		return
	}

	var re = regexp.MustCompile(`(?m).(webm|mp4|mkv|gif|flv|avi|mov|wmv|asf)`)
	for _, v := range videos {
		if label != "" {
			v.Title = label
		}

		v.Title = re.ReplaceAllString(v.Title, "")
	}

	return videos, nil
}
