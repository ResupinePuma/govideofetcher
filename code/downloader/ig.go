package downloader

import (
	"context"
	"io"
	"net/http"
	"strings"
)

type IG struct {
	SizeLimit int `yaml:"-"`
	Timeout   int `yaml:"-"`

	ReelDUrl      string "yaml:\"reel_url\""
	StoryDUrl     string "yaml:\"story_url\""
	SplashURL     string "yaml:\"splash_url\""
	SplashRequest string "yaml:\"splash_request\""

	log AbstractLogger
	ntf AbstractNotifier
}

type IGVideo struct {
	URL   string `json:"vurl"`
	Title string `json:"vtitle"`
}

type IgramWorld struct {
	Result []struct {
		VideoVersions []struct {
			Type   int    `json:"type"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
			URL    string `json:"url"`
		} `json:"video_versions"`
	} `json:"result"`
}

func (tt *IG) Init(logger AbstractLogger, notifier AbstractNotifier, opts *Opts) error {
	tt.log = logger
	tt.Timeout = opts.Timeout
	tt.SizeLimit = opts.SizeLimit
	tt.ntf = notifier
	return nil
}

func (tt *IG) httprequest(ctx context.Context, method string, url string, headers map[string]string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	return
}

func (tt *IG) Download(ctx context.Context, url string) (title string, rdr io.ReadCloser, err error) {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tt.Timeout))
	// defer cancel()

	var ttv IGVideo

	tt.ntf.Message("‍🔍 searching video")
	switch {
	case strings.Contains(url, "/reel/"):
		ttv, err = tt.getIGreel(ctx, url)
	case strings.Contains(url, "/stories/"):
		ttv, err = tt.getIGstory(ctx, url)
	}
	if err != nil {
		return
	}

	tt.ntf.Message("‍⏬ downloading video")
	res, err := tt.httprequest(ctx, http.MethodGet, ttv.URL, map[string]string{
		"User-Agent": "Mozilla/5.0 (Linux; Android 12; SM-F926B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
	}, nil)
	if err != nil {
		return
	}

	cropts := CountingReaderOpts{
		ByteLimit: tt.SizeLimit,
		FileSize:  float64(res.ContentLength),
		Notifier:  tt.ntf,
	}
	return "", NewCountingReader(res.Body, &cropts), err
}

func (tt *IG) Close() error {
	return nil
}
