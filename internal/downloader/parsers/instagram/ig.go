package instagram

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"time"

	cr "videofetcher/internal/counting_reader"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/derrors"
	v "videofetcher/internal/downloader/video"
	"videofetcher/internal/utils"

	"github.com/dop251/goja"

	"github.com/antchfx/htmlquery"
)

var (
	UA      = "Mozilla/5.0 (Linux; Android 12; SM-F926B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36"
	re      = regexp.MustCompile(`(?m)<input name="token" value="([a-zA-Z0-9=]+)" type="hidden">`)
	Logging Logger
)

type IG struct {
	SizeLimit int64
	Timeout   int
	Client    http.Client

	token        string
	lastTokenUpd time.Time
}

type IGInfo struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func NewParser(sizelim int64) *IG {
	return &IG{
		SizeLimit: sizelim,
		Client:    *http.DefaultClient,
	}
}

func (tt *IG) getToken(ctx context.Context) error {
	resp, err := utils.HTTPRequest(ctx, tt.Client, http.MethodGet, "https://snapinst.app",
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

func parseIgVideo(body string) (tv []v.Video, err error) {
	body = strings.ReplaceAll(body, `\"`, `"`)

	ps := strings.Replace(body, "return decodeURIComponent(escape(r))", `def = decodeURIComponent(escape(r)).replace(/\n/g, '').replace(/\\"/g, '"');`, -1)

	//fmt.Printf("%s", ps)
	vm := goja.New()
	vm.Set("def", "")
	_, err = vm.RunString(ps)
	if err != nil {
		return
	}
	vi := vm.Get("def")

	doc, err := htmlquery.Parse(strings.NewReader(vi.String()))
	if err != nil {
		return
	}

	t := htmlquery.Find(doc, `//div[@class="download-bottom"]/a`)
	if t == nil {
		err = derrors.ErrNotFound
		return
	}
	for _, vd := range t {
		tv = append(tv,
			*v.NewVideo("", htmlquery.SelectAttr(vd, "href"), nil),
		)
	}
	if len(tv) == 0 {
		return nil, derrors.ErrNotFound
	}

	return tv, nil
}

func (i *IG) Download(ctx *dcontext.Context) (err error) {
	ctx.Notifier().UpdTextNotify("‍🔍 searching video")

	if i.token == "" || time.Now().Sub(i.lastTokenUpd) > time.Minute*5 {
		err = i.getToken(ctx)
		if err != nil {
			return
		}
	} // if token is invalid or not updated in 5 minutes

	u := ctx.GetUrl()

	var buff bytes.Buffer
	w := multipart.NewWriter(&buff)
	w.WriteField("url", ctx.GetUrl().String())
	w.WriteField("action", "post")
	w.WriteField("token", i.token)
	w.Close()

	resp, err := utils.HTTPRequest(ctx, i.Client, http.MethodPost, "https://snapinst.app/action2.php", map[string]string{
		"User-Agent":   UA,
		"Content-Type": w.FormDataContentType(),
		"Origin":       "https://snapinst.app",
		"Referer":      "https://snapinst.app/",
	}, &buff)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("bad status code: %v", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	ttv, err := parseIgVideo(string(body))
	if err != nil {
		return
	}

	//dr := dresult.NewDownloaderResult(ctx)

	var vids []v.Video

	ctx.Notifier().UpdTextNotify("😵⏬ downloading video")
	go ctx.Notifier().StartTicker(ctx)
	for num, vid := range ttv {
		if num >= 10 {
			break
		}

		resp, err = utils.HTTPRequest(ctx, i.Client, http.MethodGet, vid.URL, map[string]string{
			"User-Agent": UA,
		}, nil)
		if err != nil {
			return
		}

		cropts := cr.CountingReaderOpts{
			ByteLimit: i.SizeLimit,
			FileSize:  float64(resp.ContentLength),
		}

		vids = append(vids,
			*v.NewVideo(u.String(), vid.URL, cr.NewCountingReader(resp.Body, &cropts)),
		)
	}

	ctx.AddResult(vids)

	close(ctx.Results())

	return err
}

func (tt *IG) Close() error {
	return nil
}
