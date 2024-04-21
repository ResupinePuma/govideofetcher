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
	TT   downloader.TTParse `yaml:"tiktok"`
	IG   downloader.IG      `yaml:"instagram"`
	YTDL downloader.YTdl    `yaml:"youtube_dl"`
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
	TT: *downloader.NewTTParse(),
	IG: downloader.IG{
		SplashURL:     "http://127.0.0.1:8050/execute",
		SplashRequest: `ZnVuY3Rpb24gbWFpbihzcGxhc2gsIGFyZ3MpCglzcGxhc2g6b25fcmVxdWVzdChmdW5jdGlvbihyZXF1ZXN0KQogICAgICAgICAgICAgICAgcmVxdWVzdDpzZXRfdGltZW91dCgxNS4wKQoJCXJlcXVlc3Q6c2V0X2h0dHAyX2VuYWJsZWQodHJ1ZSkKCQlyZXF1ZXN0LmhlYWRlcnNbInVzZXItYWdlbnQiXSA9ICJNb3ppbGxhLzUuMCAoTGludXg7IEFuZHJvaWQgMTI7IFNNLUY5MjZCKSBBcHBsZVdlYktpdC81MzcuMzYgKEtIVE1MLCBsaWtlIEdlY2tvKSBDaHJvbWUvMTA3LjAuMC4wIFNhZmFyaS81MzcuMzYiCgllbmQpCiAgc3BsYXNoLmltYWdlc19lbmFibGVkID0gZmFsc2UKCWFzc2VydChzcGxhc2g6Z28oYXJncy51cmwpKQogIGFzc2VydChzcGxhc2g6d2FpdCgwLjUpKQoJc3BsYXNoOnJ1bmpzJ2RvY3VtZW50LmdldEVsZW1lbnRzQnlOYW1lKCJxIilbMF0udmFsdWU9IiVzIjtkb2N1bWVudC5nZXRFbGVtZW50c0J5Q2xhc3NOYW1lKCJidG4gYnRuLWRlZmF1bHQiKVswXS5jbGljaygpOycKICB3aGlsZSBub3Qgc3BsYXNoOnNlbGVjdCIuZG93bmxvYWQtYm94IiBkbyBhc3NlcnQoc3BsYXNoOndhaXQoMSkpIGVuZAogIGxvY2FsIHZ1cmwgPSBzcGxhc2g6ZXZhbGpzJ0FycmF5LnByb3RvdHlwZS5zbGljZS5jYWxsKGRvY3VtZW50LmdldEVsZW1lbnRzQnlDbGFzc05hbWUoImRvd25sb2FkLWl0ZW1zX19idG4iKSkuc2xpY2UoLTEpWzBdLmdldEVsZW1lbnRzQnlUYWdOYW1lKCJhIilbMF0uaHJlZjsnCiAgcmV0dXJuIHt7IHZ1cmwgPSB2dXJsfX0KZW5k`,
		ReelDUrl:      "https://saveinsta.app/en1",
		StoryDUrl:     "https://igram.world/api/ig/story?url=%s",
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

	//cfg.TT.SplashRequest = base64.StdEncoding.EncodeToString([]byte(cfg.TT.SplashRequest))
	cfg.IG.SplashRequest = base64.StdEncoding.EncodeToString([]byte(cfg.IG.SplashRequest))
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
	// tc, err := base64.StdEncoding.DecodeString(cfg.TT.SplashRequest)
	// if err != nil {
	// 	return
	// }
	// cfg.TT.SplashRequest = string(tc)
	cfg.TT = *downloader.NewTTParse()

	// ti, err := base64.StdEncoding.DecodeString(cfg.IG.SplashRequest)
	// if err != nil {
	// 	return
	// }
	// cfg.IG.SplashRequest = string(ti)
	return
}
