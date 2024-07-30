package reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	UserAgent     = "TelegramBot:YetAnotherFileFetcher/0.0.1"
	PostApiPoint  = "https://api.reddit.com/api/info/?id=t3_"
	TokenApiPoint = "https://www.reddit.com/api/v1/access_token"

	grantType = "grant_type=client_credentials&duration=permanent"
)

type authData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresUntil int64  `json:"expires_in"`
}

func findId(url string) (id, t string) {
	myExp, _ := regexp.Compile(`(?:^.+?)(?:reddit.com\/r)(?:\/[\w\d]+)(?:\/)(?P<type>[\w\d]+)(?:\/)(?P<id>[\w\d]*)`)
	match := myExp.FindStringSubmatch(url)
	for i, name := range myExp.SubexpNames() {
		if i != 0 {
			switch name {
			case "id":
				id = match[i]
			case "type":
				t = match[i]
			}
		}
	}

	return
}

func (ra *Reddit) getToken(ctx context.Context) error {
	if ra.auth != nil && time.Now().UTC().Unix() < ra.auth.ExpiresUntil {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		TokenApiPoint,
		strings.NewReader(grantType),
	)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(ra.ClientID, ra.Secret)

	resp, err := ra.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth err. got resp: %s", data)
	}

	err = json.NewDecoder(resp.Body).Decode(&ra.auth)
	if err != nil {
		return err
	}

	if ra.auth.ExpiresUntil != 0 {
		ra.auth.ExpiresUntil = time.Now().UTC().Unix() + ra.auth.ExpiresUntil
	}

	return nil
}

func (ra *Reddit) GetHeaders(ctx context.Context) (http.Header, error) {
	err := ra.getToken(ctx)
	if err != nil {
		return nil, err
	}

	hdrs := http.Header{}
	hdrs.Set("User-Agent", UserAgent)
	hdrs.Set("Authorization", fmt.Sprintf("bearer: %s", ra.auth.AccessToken))

	return hdrs, nil
}

func (ra *Reddit) GetHLSUrl(ctx context.Context, url string) (hlsu, title string, err error) {
	err = ra.getToken(ctx)
	if err != nil {
		return "", "", err
	}

repeat:
	id, ltype := findId(url)
	if id == "" {
		return "", "", fmt.Errorf("empty id")
	}

	if ltype == "" {
		return "", "", fmt.Errorf("wrong type")
	}

	if ltype != "comments" {
		req, err := http.NewRequestWithContext(ctx,
			http.MethodGet,
			url,
			nil,
		)
		if err != nil {
			return "", "", err
		}
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set(`Autorization`, fmt.Sprintf("bearer: %s", ra.auth.AccessToken))
		resp, err := ra.client.Do(req)
		if err != nil {
			return "", "", err
		}
		if resp.StatusCode != http.StatusOK {
			data, _ := io.ReadAll(resp.Body)
			return "", "", fmt.Errorf("auth err. got resp: %s", data)
		}

		url = resp.Request.URL.String()
		goto repeat
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		PostApiPoint+id,
		strings.NewReader(grantType),
	)
	if err != nil {
		return "", "", err
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set(`Autorization`, fmt.Sprintf("bearer: %s", ra.auth.AccessToken))

	resp, err := ra.client.Do(req)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("auth err. got resp: %s", data)
	}

	dec := json.NewDecoder(resp.Body)

	for {
		t, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return hlsu, title, nil // Конец файла
			}
			return "", "", err
		}

		if key, ok := t.(string); ok {
			// Читаем следующее значение
			if key == "title" {
				if err := dec.Decode(&title); err != nil {
					return "", "", err
				}
			}

			if key == "hls_url" {
				if err := dec.Decode(&hlsu); err != nil {
					return "", "", err
				}
			}

			continue
		}

	}
}
