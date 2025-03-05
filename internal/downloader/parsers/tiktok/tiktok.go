package tiktok

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

	cr "videofetcher/internal/counting_reader"
	"videofetcher/internal/downloader/dcontext"
	v "videofetcher/internal/downloader/video"
	"videofetcher/internal/utils"
)

var (
	UA  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
	REF = "https://snaptik.app/"

	re   = regexp.MustCompile(`(?m)<input name="token" value="([a-zA-Z0-9=]+)" type="hidden">`)
	jsr  = regexp.MustCompile(`(?m)<script>(.*)<\/script>`)
	er   = regexp.MustCompile(`(?m)parent.document.getElementById("alert").innerHTML = '(.*)';`)
	aaaa = regexp.MustCompile(`(?m)\$\("#download"\)\.innerHTML = "(.*<\/div>)";`)

	Logging Logger
)

func parseVideo(body string) (tv v.Video, err error) {
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
	if as == nil {
		err = fmt.Errorf("can't find url")
		return
	}
	for _, v := range as {
		tv.URL = htmlquery.SelectAttr(v, "href")
		break
	}

	return
}

type TikTok struct {
	SizeLimit int64

	Client http.Client

	token        string
	lastTokenUpd time.Time
}

func NewParser(sizelim int64) *TikTok {
	return &TikTok{
		SizeLimit: sizelim,
		Client:    *http.DefaultClient,
	}
}

func (tt *TikTok) getToken(ctx context.Context) error {
	resp, err := utils.HTTPRequest(ctx, tt.Client, http.MethodGet, REF,
		map[string]string{
			"User-Agent": UA,
		},
		nil)
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
	tt.lastTokenUpd = time.Now()
	return nil
}

func (tt *TikTok) getJsData(ctx context.Context, u string, headers map[string]string) (js string, err error) {
	data := url.Values{}
	data.Set("url", u)
	data.Set("token", tt.token)

	resp, err := utils.HTTPRequest(ctx, tt.Client, http.MethodPost, "https://snaptik.app/abc2.php", headers, strings.NewReader(data.Encode()))
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

func (tt *TikTok) Download(ctx *dcontext.Context) (err error) {
	if tt.token == "" || time.Now().Sub(tt.lastTokenUpd) > time.Minute*5 {
		err = tt.getToken(ctx)
		if err != nil {
			return
		}
	} // if token is invalid or not updated in 5 minutes

	u := ctx.GetUrl()

	headers := map[string]string{
		"User-Agent":   UA,
		"Content-Type": "application/x-www-form-urlencoded",
		"Origin":       REF,
		"Referer":      REF,
	}

	ctx.Notifier().UpdTextNotify("‍🔍 searching video")
	ps, err := tt.getJsData(ctx, u.String(), headers)
	if err != nil {
		return
	}

	ctx.Notifier().UpdTextNotify("‍🔍 getting video info")
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

	//dr := dresult.NewDownloaderResult(ctx)

	ctx.Notifier().UpdTextNotify("‍⏬ downloading video")
	go ctx.Notifier().StartTicker(ctx)
	resp, err := utils.HTTPRequest(ctx, tt.Client, http.MethodGet, tv.URL, headers, nil)
	if err != nil {
		return
	}

	cropts := cr.CountingReaderOpts{
		ByteLimit: tt.SizeLimit,
		FileSize:  float64(resp.ContentLength),
	}

	ctx.AddResult([]v.Video{*v.NewVideo(tv.Title, u.String(), cr.NewCountingReader(resp.Body, &cropts))})

	close(ctx.Results())

	return nil
}

func (tt *TikTok) Close() error { return nil }
