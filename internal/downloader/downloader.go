package downloader

import (
	"regexp"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/options"
	"videofetcher/internal/downloader/parsers/instagram"
	"videofetcher/internal/downloader/parsers/tiktok"
	"videofetcher/internal/downloader/parsers/ytdl"
	"videofetcher/internal/downloader/video"
	"videofetcher/internal/utils"
)

var (
	ParserTikTok  = "tt"
	ParserIG      = "ig"
	ParserYTMusic = "mytdl"
	ParserDefault = "any"
)

type Downloader struct {
	sizelimit int64
	timeout   int64

	notifier AbstractNotifier
	parsers  map[string]AbstractDownloader
}

func NewDownloader(opts options.DownloaderOpts) *Downloader {
	d := &Downloader{
		sizelimit: opts.SizeLimit, //100Mb
		timeout:   opts.Timeout,
	}
	d.parsers = map[string]AbstractDownloader{
		ParserTikTok:  tiktok.NewParser(d.sizelimit),
		ParserIG:      instagram.NewParser(d.sizelimit),
		ParserDefault: ytdl.NewParser(d.sizelimit, &opts.YTDL),
	}
	return d
}

func (d *Downloader) SetNotifier(n AbstractNotifier) {
	d.notifier = n
}

func (d *Downloader) Download(ctx dcontext.Context, text string) (res []video.Video, err error) {

	url, label, err := utils.ExtractUrlAndText(text)
	if err != nil {
		return
	}

	var dwn AbstractDownloader
	switch url.Hostname() {
	case "tiktok.com", "vt.tiktok.com":
		dwn = d.parsers[ParserTikTok]
	case "instagram.com", "www.instagram.com":
		dwn = d.parsers[ParserIG]
	default:
		dwn = d.parsers[ParserDefault]
	}

	videos, err := dwn.Download(ctx, url.String())
	if err != nil {
		return
	}

	var re = regexp.MustCompile(`(?m).(webm|mp4|mkv|gif|flv|avi|mov|wmv|asf)`)
	for i, v := range videos {
		if label != "" {
			v.Title = label
		}

		v.Title = re.ReplaceAllString(v.Title, "")
		videos[i] = v
	}

	return videos, nil
}
