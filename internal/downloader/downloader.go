package downloader

import (
	"net/http"
	"net/url"
	"sync"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/media"
	"videofetcher/internal/downloader/options"
	"videofetcher/internal/downloader/parsers/instagram"
	"videofetcher/internal/downloader/parsers/tiktok"
	"videofetcher/internal/downloader/parsers/ytdl"
	"videofetcher/internal/proxiedHTTP"
)

var (
	ParserTikTok  = "tt"
	ParserIG      = "ig"
	ParserYTMusic = "mytdl"
	//ParserReddit  = "rddt"
	ParserDefault = "any"

	Logger iLogger
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
	var c http.Client
	if opts.DownloaderProxy != "" {
		c = proxiedHTTP.NewProxiedHTTPClient(opts.DownloaderProxy)
	} else {
		c = *http.DefaultClient
	}

	opts.YTDL.Proxies = opts.DownloaderProxy

	d.parsers = map[string]AbstractDownloader{
		ParserTikTok:  tiktok.NewParser(d.sizelimit, c),
		ParserIG:      instagram.NewParser(d.sizelimit, c),
		ParserYTMusic: ytdl.NewParserAudio(d.sizelimit, opts.YTDL),
		ParserDefault: ytdl.NewParser(d.sizelimit, opts.YTDL),
		//ParserReddit:  reddit.NewParser(d.sizelimit, &opts.Reddit, &opts.YTDL),
	}
	return d
}

func (d *Downloader) SetNotifier(n AbstractNotifier) {
	d.notifier = n
}

func (d *Downloader) Download(ctx *dcontext.Context, u *url.URL) (res []media.Media, err error) {
	ctx, span := dcontext.NewTracerContext(ctx, "download")
	defer span.End()

	var dwn AbstractDownloader
	switch u.Hostname() {
	case "www.tiktok.com", "tiktok.com", "vt.tiktok.com":
		dwn = d.parsers[ParserTikTok]
	case "music.youtube.com", "soundcloud.com":
		dwn = d.parsers[ParserYTMusic]
	case "instagram.com", "www.instagram.com":
		dwn = d.parsers[ParserIG]
	// case "www.reddit.com", "old.reddit.com", "reddit.com", "redd.it", "v.redd.it":
	// 	dwn = d.parsers[ParserReddit]
	default:
		dwn = d.parsers[ParserDefault]
	}

	var errch chan error = make(chan error)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := dwn.Download(ctx)
		if err != nil {
			errch <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for rr := range ctx.Results() {
			res = append(res, rr...)
		}
	}()

	go func() {
		wg.Wait()
		close(errch)
	}()

	for e := range errch {
		if e != nil {
			return nil, e
		}
	}

	return res, nil
}
