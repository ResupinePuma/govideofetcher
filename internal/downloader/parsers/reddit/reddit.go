package reddit

import (
	"net/http"
	"net/url"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/options"
	"videofetcher/internal/downloader/parsers/ytdl"
	"videofetcher/internal/downloader/video"
)

type Reddit struct {
	ClientID string
	Secret   string

	auth    *authData
	sizelim int64

	client *http.Client
	dwn    *ytdl.YtDl
}

func NewParser(sizelim int64, ropts *options.RedditOptions, opts *options.YTDLOptions) *Reddit {
	yt := Reddit{
		sizelim:  sizelim,
		ClientID: ropts.ID,
		Secret:   ropts.Secret,
		dwn: &ytdl.YtDl{
			SizeLimit:  sizelim,
			Format:     opts.Format,
			Executable: opts.Executable,
		},
		client: &http.Client{},
	}
	return &yt
}

func (rd *Reddit) Download(ctx dcontext.Context, u *url.URL) ([]video.Video, error) {
	headers, err := rd.GetHeaders(&ctx)
	if err != nil {
		return nil, err
	}

	rd.dwn.Headers = headers

	hslurl, title, err := rd.GetHLSUrl(&ctx, u.String())
	if err != nil {
		return nil, err
	}

	hu, err := url.Parse(hslurl)
	if err != nil {
		return nil, err
	}

	videos, err := rd.dwn.Download(ctx, hu)
	for i, v := range videos {
		v.Title = title
		videos[i] = v
	}
	return videos, err
}

func (rd *Reddit) Close() error {
	return rd.dwn.Close()
}
