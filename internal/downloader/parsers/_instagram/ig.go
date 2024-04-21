package instagram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/antchfx/htmlquery"
)

type IG struct {
	SizeLimit int `yaml:"-"`
	Timeout   int `yaml:"-"`

	log AbstractLogger
	ntf AbstractNotifier
}

type IGVideo struct {
	URL   string `json:"vurl"`
	Title string `json:"vtitle"`
}

type IgramInfo struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func parseIgVideo(body string) (tv IGVideo, err error) {
	body = strings.ReplaceAll(body, `\"`, `"`)

	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	t := htmlquery.Find(doc, `//div[@class="download-items__btn"]/a`)
	if t == nil {
		err = fmt.Errorf("can't find title")
		return
	}
	for _, v := range t {
		tv.URL = htmlquery.SelectAttr(v, "href")
		return
	}
	return tv, fmt.Errorf("no video")
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

func (i *IG) Download(ctx context.Context, u string) (title string, rdr io.ReadCloser, err error) {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tt.Timeout))
	// defer cancel()

	var ttv IGVideo

	i.ntf.Message("‍🔍 searching video")

	data := url.Values{}
	data.Set("q", u)
	data.Set("t", "media")
	data.Set("lang", "en")

	resp, err := i.httprequest(ctx, http.MethodPost, "https://v3.saveig.app/api/ajaxSearch", map[string]string{
		"User-Agent":   "Mozilla/5.0 (Linux; Android 12; SM-F926B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
		"Content-Type": "application/x-www-form-urlencoded",
		"Origin":       "https://saveig.app",
		"Referer":      "https://saveig.app/",
	}, strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("bad status code: %v", resp.StatusCode)
		return
	}

	tmp := IgramInfo{}
	err = json.NewDecoder(resp.Body).Decode(&tmp)
	if err != nil {
		return
	}
	if tmp.Status != "ok" {
		err = errors.New("can't find video")
		return
	}

	ttv, err = parseIgVideo(tmp.Data)
	if err != nil {
		return
	}

	i.ntf.Message("‍⏬ downloading video")
	res, err := i.httprequest(ctx, http.MethodGet, ttv.URL, map[string]string{
		"User-Agent": "Mozilla/5.0 (Linux; Android 12; SM-F926B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
	}, nil)
	if err != nil {
		return
	}

	cropts := CountingReaderOpts{
		ByteLimit: i.SizeLimit,
		FileSize:  float64(res.ContentLength),
		Notifier:  i.ntf,
	}
	return "", NewCountingReader(res.Body, &cropts), err
}

func (tt *IG) Close() error {
	return nil
}
