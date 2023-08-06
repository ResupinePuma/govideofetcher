package downloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
)

func (i *IG) getIGstory(ctx context.Context, u string) (t IGVideo, err error) {
	u = url.QueryEscape(u)
	res, err := i.httprequest(ctx, http.MethodGet, fmt.Sprintf(i.StoryDUrl, u), map[string]string{
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
