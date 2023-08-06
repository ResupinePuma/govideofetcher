package main

import (
	"encoding/base64"
	"errors"
	"os"
	"videofetcher/downloader"

	"gopkg.in/yaml.v2"
)

var ErrCfg = errors.New("error reading config.yml")

type Config struct {
	Base struct {
		Token     string `yaml:"tg_token"`
		Timeout   int    `yaml:"timeout"`
		SizeLimit int    `yaml:"size_limit"`
	} `yaml:"base"`
	TT   downloader.TikTok `yaml:"tiktok"`
	IG   downloader.IG     `yaml:"instagram"`
	YTDL downloader.YTdl   `yaml:"youtube_dl"`
}

var cfg = Config{
	Base: struct {
		Token     string `yaml:"tg_token"`
		Timeout   int    "yaml:\"timeout\""
		SizeLimit int    "yaml:\"size_limit\""
	}{
		Token:     "0000000000:11111111111111111111111111111111111",
		Timeout:   30,
		SizeLimit: 50 * 1024 * 1024,
	},
	TT: downloader.TikTok{
		SplashURL:     "http://127.0.0.1:8050/execute",
		SplashRequest: `ZnVuY3Rpb24gbWFpbihzcGxhc2gsIGFyZ3MpCglzcGxhc2g6b25fcmVxdWVzdChmdW5jdGlvbihyZXF1ZXN0KQoJCXJlcXVlc3Q6c2V0X2h0dHAyX2VuYWJsZWQodHJ1ZSkKCQlyZXF1ZXN0LmhlYWRlcnNbInVzZXItYWdlbnQiXSA9ICJNb3ppbGxhLzUuMCAoTGludXg7IEFuZHJvaWQgMTI7IFNNLUY5MjZCKSBBcHBsZVdlYktpdC81MzcuMzYgKEtIVE1MLCBsaWtlIEdlY2tvKSBDaHJvbWUvMTA3LjAuMC4wIFNhZmFyaS81MzcuMzYiCgllbmQpCglzcGxhc2guaW1hZ2VzX2VuYWJsZWQgPSBmYWxzZQoJc3BsYXNoOmdvKGFyZ3MudXJsKQoJc3BsYXNoOnJ1bmpzJ2RvY3VtZW50LmdldEVsZW1lbnRzQnlOYW1lKCJ1cmwiKVswXS52YWx1ZT0iJXMiO2RvY3VtZW50LmdldEVsZW1lbnRzQnlUYWdOYW1lKCJmb3JtIilbMF0uc3VibWl0KCk7JwoJd2hpbGUgbm90IHNwbGFzaDpzZWxlY3QiI2Rvd25sb2FkLWJsb2NrIiBkbyBhc3NlcnQoc3BsYXNoOndhaXQoMSkpIGVuZAoJbG9jYWwgdnVybCA9IHNwbGFzaDpldmFsanMnZG9jdW1lbnQuZ2V0RWxlbWVudHNCeUNsYXNzTmFtZSgiYWJ1dHRvbnMgbWItMCIpWzBdLmNoaWxkcmVuWzBdLmhyZWYnCglsb2NhbCB0aXRsZSA9IHNwbGFzaDpldmFsanMnZG9jdW1lbnQuZ2V0RWxlbWVudHNCeUNsYXNzTmFtZSgidmlkZW90aWttYXRlLW1pZGRsZSBjZW50ZXIiKVswXS5jaGlsZHJlblsxXS5pbm5lclRleHQnCglyZXR1cm4geyB7IHRpdGxlID0gdGl0bGUsIHZ1cmwgPSB2dXJsIH0gfQplbmQ=`,
		TTUrl:         "https://tikmate.online/?lang=nl",
	},
	IG: downloader.IG{
		IGUrl: "https://igram.world/api/ig/story?url=%s",
	},
	YTDL: downloader.YTdl{
		Format: "18/17,bestvideo[height<=720][ext=mp4]+worstaudio,(mp4)[ext=mp4][vcodec^=h26],worst[width>=480][ext=mp4],worst[ext=mp4]",
	},
}

func NewDefaultConfig() (err error) {
	err = os.Mkdir("configs", 0644)
	if err != nil {
		return
	}
	f, err := os.Create("configs/config.yml")
	if err != nil {
		return
	}
	defer f.Close()

	cfg.TT.SplashRequest = base64.StdEncoding.EncodeToString([]byte(cfg.TT.SplashRequest))
	encoder := yaml.NewEncoder(f)
	encoder.Encode(cfg)
	if err != nil {
		return
	}
	return
}

func NewConfig() (cfg *Config, err error) {
	f, err := os.Open("configs/config.yml")
	if err != nil {
		return nil, ErrCfg
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return
	}
	tc, err := base64.StdEncoding.DecodeString(cfg.TT.SplashRequest)
	if err != nil {
		return
	}
	cfg.TT.SplashRequest = string(tc)
	return
}
