package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/dop251/goja"
)

var (
	UA  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
	REF = "https://snaptik.app/"

	re   = regexp.MustCompile(`(?m)<input name="token" value="([a-zA-Z0-9=]+)" type="hidden">`)
	jsr  = regexp.MustCompile(`(?m)<script>(.*)<\/script>`)
	er   = regexp.MustCompile(`(?m)parent.document.getElementById("alert").innerHTML = '(.*)';`)
	aaaa = regexp.MustCompile(`(?m)\$\("#download"\)\.innerHTML = "(.*<\/div>)";`)

	client = http.Client{}
)

func valError(body string) error {
	err := er.FindString(body)
	if err != "" {
		return fmt.Errorf(err)
	}
	return nil
}

func parseVideo(body string) (tv TTVideo, err error) {
	body = strings.ReplaceAll(body, `\"`, `"`)

	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	t := htmlquery.FindOne(doc, `//div[@class="video-title"]`)
	if t == nil {
		err = fmt.Errorf("can't find title")
		return
	}
	tv.Title = htmlquery.InnerText(t)

	as := htmlquery.Find(doc, `//div[@class="video-links"]/a`)
	if t == nil {
		err = fmt.Errorf("can't find url")
		return
	}
	for _, v := range as {
		tv.URL = htmlquery.SelectAttr(v, "href")
		break
	}

	return
}

type TTVideo struct {
	URL   string `json:"vurl"`
	Title string `json:"title"`
}

type TTParse struct {
	SizeLimit int `yaml:"-"`

	token string

	logger AbstractLogger
	notify AbstractNotifier

	client http.Client
	lastR  time.Time

	vm *goja.Runtime
}

func NewTTParse() *TTParse {
	return &TTParse{
		client: http.Client{},
		vm:     nil,
	}
}

func (tt *TTParse) Init(loggger AbstractLogger, notifier AbstractNotifier, opts *Opts) error {
	tt.logger = loggger
	tt.notify = notifier
	tt.SizeLimit = opts.SizeLimit
	return nil
}

func (tt *TTParse) httprequest(ctx context.Context, method string, url string, headers map[string]string, body io.Reader) (resp *http.Response, err error) {
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

func (tt *TTParse) getToken(ctx context.Context) error {
	resp, err := tt.httprequest(ctx, http.MethodGet, REF, map[string]string{
		"User-Agent": UA,
	}, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status code: %v", resp.StatusCode)
	}

	page, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	tokens := re.FindSubmatch(page)
	if tokens == nil {
		return fmt.Errorf("token not found")
	}

	tt.token = string(tokens[len(tokens)-1])
	tt.lastR = time.Now()
	return nil
}

func (tt *TTParse) getJsData(ctx context.Context, u string, headers map[string]string) (js string, err error) {
	data := url.Values{}
	data.Set("url", u)
	data.Set("token", tt.token)

	resp, err := tt.httprequest(ctx, http.MethodPost, "https://snaptik.app/abc2.php", headers, strings.NewReader(data.Encode()))
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("bad status code: %v", resp.StatusCode)
		return
	}

	pb, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	js = strings.Replace(string(pb), "return decodeURIComponent(escape(r))", "def = decodeURIComponent(escape(r))", -1)
	return
}

func (tt *TTParse) Download(ctx context.Context, u string) (title string, v io.ReadCloser, err error) {
	if tt.token == "" || time.Now().Sub(tt.lastR) > time.Minute*5 {
		err = tt.getToken(ctx)
		if err != nil {
			return
		}
	}

	headers := map[string]string{
		"User-Agent":   UA,
		"Content-Type": "application/x-www-form-urlencoded",
		"Origin":       REF,
		"Referer":      REF,
	}

	tt.notify.Message("‍🔍 searching video")
	ps, err := tt.getJsData(ctx, u, headers)
	if err != nil {
		return
	}

	tt.notify.Message("‍🔍 getting video info")
	vm := goja.New()
	vm.Set("def", "")
	_, err = vm.RunString(ps)
	if err != nil {
		return
	}
	vi := vm.Get("def")
	ta := aaaa.FindStringSubmatch(vi.String())
	if len(ta) < 2 {
		err = fmt.Errorf("video not found")
		return
	}

	tv, err := parseVideo(ta[1])
	if err != nil {
		return
	}

	tt.notify.Message("‍⏬ downloading video")
	resp, err := tt.httprequest(ctx, http.MethodGet, tv.URL, headers, nil)
	if err != nil {
		return
	}

	cropts := CountingReaderOpts{
		ByteLimit: tt.SizeLimit,
		FileSize:  float64(resp.ContentLength),
		Notifier:  tt.notify,
	}
	return tv.Title, NewCountingReader(resp.Body, &cropts), err
}

func (tt *TTParse) Close() error { return nil }
