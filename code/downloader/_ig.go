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

type IGVideo struct {
	URL   string `json:"vurl"`
}

type IG struct {
	SizeLimit int `yaml:"-"`
	Timeout   int `yaml:"-"`

	IGUrl         string "yaml:\"ig_url\""
	SplashURL     string "yaml:\"splash_url\""
	SplashRequest string "yaml:\"splash_request\""

	log AbstractLogger
	ntf AbstractNotifier
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

func (tt *IG) getIGvideo(ctx context.Context, url string) (t IGVideo, err error) {
	reqJson := map[string]string{
		"url":        tt.IGUrl,
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

	tmp := map[string]IGVideo{}
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
		"User-Agent":"Mozilla/5.0 (Linux; Android 12; SM-F926B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
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
