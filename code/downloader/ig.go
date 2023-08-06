package downloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
)

type IGVideo struct {
	URL string `json:"vurl"`
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

type IG struct {
	SizeLimit int `yaml:"-"`
	Timeout   int `yaml:"-"`

	IGUrl string "yaml:\"ig_url\""
	log   AbstractLogger
	ntf   AbstractNotifier
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

func (i *IG) getIGvideo(ctx context.Context, u string) (t IGVideo, err error) {
	u = url.QueryEscape(u)
	res, err := i.httprequest(ctx, http.MethodGet, fmt.Sprintf(i.IGUrl, u), map[string]string{
		"User-Agent":   "Mozilla/5.0 (Linux; Android 12; SM-F926B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
		"Content-Type": "application/json",
	}, nil)
	if err != nil {
		return
	}

	tmp := IgramWorld{}
	err = json.NewDecoder(res.Body).Decode(&tmp)
	if err != nil {
		return
	}
	if len(tmp.Result) == 0 {
		err = errors.New("can't find video")
		return
	}
	vres := tmp.Result[0]
	sort.Slice(vres.VideoVersions, func(i, j int) bool {
		return vres.VideoVersions[i].Type < vres.VideoVersions[j].Type
	})

	t.URL = tmp.Result[0].VideoVersions[0].URL
	return t, nil
}

func (tt *IG) Download(ctx context.Context, url string) (title string, rdr io.ReadCloser, err error) {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tt.Timeout))
	// defer cancel()
	tt.ntf.Message("‍🔍 searching video")
	ttv, err := tt.getIGvideo(ctx, url)
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
