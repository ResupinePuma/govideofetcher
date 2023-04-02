package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type TTVideo struct {
	Url   string `json:"vurl"`
	Title string `json:"title"`
}

type TikTok struct {
	SizeLimit int `yaml:"-"`
	Timeout   int `yaml:"-"`

	TTUrl         string "yaml:\"tt_url\""
	SplashURL     string "yaml:\"splash_url\""
	SplashRequest string "yaml:\"splash_request\""

	log AbstractLogger
	ntf AbstractNotifier
}

func (tt *TikTok) Init(logger AbstractLogger, notifier AbstractNotifier, opts *DownloaderOpts) error {
	tt.log = logger
	tt.Timeout = opts.Timeout
	tt.SizeLimit = opts.SizeLimit
	tt.ntf = notifier
	return nil
}

func (tt *TikTok) httprequest(ctx context.Context, method string, url string, headers map[string]string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
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

func (tt *TikTok) getTTvideo(ctx context.Context, url string) (t TTVideo, err error) {
	reqJson := map[string]string{
		"url":        tt.TTUrl,
		"lua_source": fmt.Sprintf(tt.SplashRequest, url),
	}
	body, err := json.Marshal(reqJson)
	if err != nil {
		return
	}
	res, err := tt.httprequest(ctx, http.MethodPost, tt.SplashURL, map[string]string{
		"Content-Type": "application/json",
	}, bytes.NewReader(body))
	if err != nil {
		return
	}

	tmp := map[string]TTVideo{}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&tmp)
	if err != nil {
		return
	}
	if len(tmp) == 0 {
		err = errors.New("can't find video")
		return
	}
	return tmp["1"], nil
}

func (tt *TikTok) Download(ctx context.Context, url string) (title string, rdr io.ReadCloser, err error) {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tt.Timeout))
	// defer cancel()
	tt.ntf.Message("‍🔍 searching video")
	ttv, err := tt.getTTvideo(ctx, url)
	if err != nil {
		return
	}

	tt.ntf.Message("‍⏬ downloading video")
	res, err := tt.httprequest(ctx, http.MethodGet, ttv.Url, map[string]string{}, nil)
	if err != nil {
		return
	}

	cropts := CountingReaderOpts{
		ByteLimit: tt.SizeLimit,
		FileSize:  float64(res.ContentLength),
		Notifier:  tt.ntf,
	}
	return ttv.Title, NewCountingReader(res.Body, &cropts), err
}

func (tt *TikTok) Close() error {
	return nil
}
