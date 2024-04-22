package instagram

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	cr "videofetcher/internal/counting_reader"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/derrors"
	v "videofetcher/internal/downloader/video"
	"videofetcher/internal/utils"

	"github.com/antchfx/htmlquery"
)

var (
	UA = "Mozilla/5.0 (Linux; Android 12; SM-F926B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36"

	Logger iLogger
)

type IG struct {
	SizeLimit int64
	Timeout   int
}

type IGInfo struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func NewParser(sizelim int64) *IG {
	return &IG{
		SizeLimit: sizelim,
	}
}

func parseIgVideo(body string) (tv []v.Video, err error) {
	body = strings.ReplaceAll(body, `\"`, `"`)

	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	t := htmlquery.Find(doc, `//div[@class="download-items__btn"]/a`)
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

func (i *IG) Download(ctx dcontext.Context, u string) (vids []v.Video, err error) {
	ctx.Notifier().UpdTextNotify("‍🔍 searching video")

	resp, err := utils.HTTPRequest(&ctx, http.MethodPost, "https://v3.saveig.app/api/ajaxSearch", map[string]string{
		"User-Agent":   UA,
		"Content-Type": "application/x-www-form-urlencoded",
		"Origin":       "https://saveig.app",
		"Referer":      "https://saveig.app/",
	}, strings.NewReader(
		utils.GenerateQuery(
			map[string]string{
				"q":    u,
				"t":    "media",
				"lang": "en",
			},
		)))
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("bad status code: %v", resp.StatusCode)
		return
	}

	tmp := IGInfo{}
	err = json.NewDecoder(resp.Body).Decode(&tmp)
	if err != nil {
		return
	}
	if tmp.Status != "ok" {
		err = errors.New("can't find video")
		return
	}

	ttv, err := parseIgVideo(tmp.Data)
	if err != nil {
		return
	}

	ctx.Notifier().UpdTextNotify("‍⏬ downloading video")
	for _, vid := range ttv {
		resp, err = utils.HTTPRequest(&ctx, http.MethodGet, vid.URL, map[string]string{
			"User-Agent": UA,
		}, nil)
		if err != nil {
			return
		}

		cropts := cr.CountingReaderOpts{
			ByteLimit: i.SizeLimit,
			FileSize:  float64(resp.ContentLength),
			Notifier:  ctx.Notifier(),
		}

		vids = append(vids,
			*v.NewVideo(u, vid.URL, cr.NewCountingReader(resp.Body, &cropts)),
		)
	}

	return vids, err
}

func (tt *IG) Close() error {
	return nil
}
