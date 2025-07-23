package instagram

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"
	"videofetcher/internal/counting_reader"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/media"
	"videofetcher/internal/notice"
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

type IGVIdeo struct {
	Status string `json:"status"`
	Data   []struct {
		Title       string `json:"title"`
		Thumbnail   string `json:"thumbnail"`
		DownloadURL string `json:"downloadUrl"`
		VideoURL    string `json:"videoUrl"`
	} `json:"data"`
}

func NewParser(sizelim int64, c http.Client) *IG {
	return &IG{
		SizeLimit: sizelim,
		Client:    c,
	}
}

func (i *IG) Download(ctx *dcontext.Context) (err error) {
	ctx.Notifier().UpdTextNotify("‍🔍 searching media")

	u := ctx.GetUrl()

	data := url.Values{}
	data.Set("url", u.String())

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		"https://snapins.ai/action.php", bytes.NewReader([]byte(data.Encode())))
	if err != nil {
		return err
	}

	for k, v := range map[string]string{
		"User-Agent":   UA,
		"Content-Type": "application/x-www-form-urlencoded",
		"Origin":       "https://snapins.ai",
		"Referer":      "https://snapins.ai",
	} {
		req.Header.Add(k, v)
	}

	resp, err := i.Client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		err = notice.ErrInvalidResponse
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var IGVid IGVIdeo
	err = json.Unmarshal(body, &IGVid)
	if err != nil {
		return
	}

	if len(IGVid.Data) == 0 {
		return notice.ErrNotFound
	}

	var vids []media.Media

	ctx.Notifier().UpdTextNotify(notice.TranslateNotice(notice.NoticeDownloadingMedia, ctx.GetLang()))
	go ctx.Notifier().StartTicker(ctx)
	for num, vid := range IGVid.Data {
		if num >= 10 {
			break
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, vid.VideoURL, nil)
		if err != nil {
			return err
		}

		for k, v := range map[string]string{
			"User-Agent": UA,
		} {
			req.Header.Add(k, v)
		}

		resp, err := i.Client.Do(req)
		if err != nil {
			return err
		}

		cropts := counting_reader.CountingReaderOpts{
			ByteLimit: i.SizeLimit,
			FileSize:  float64(resp.ContentLength),
		}

		vn := md5.Sum([]byte(vid.VideoURL))
		vids = append(vids,
			media.NewVideo(hex.EncodeToString(vn[:])+".mp4", vid.Title, vid.VideoURL, counting_reader.NewCountingReader(resp.Body, &cropts)),
		)
	}

	ctx.AddResult(vids)

	close(ctx.Results())

	return err
}

func (tt *IG) Close() error {
	return nil
}
