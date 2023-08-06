package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (tt *IG) getIGreel(ctx context.Context, url string) (t IGVideo, err error) {
	reqJson := map[string]string{
		"url":        tt.ReelDUrl,
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
