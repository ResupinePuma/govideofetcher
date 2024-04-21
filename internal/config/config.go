package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v2"
)

var ErrCfg = errors.New("error reading config.yml")

type Config struct {
	Base struct {
		Token     string `yaml:"tg_token"`
		Debug     bool   `yaml:"debug"`
		SizeLimit int64  `yaml:"size_limit"`
		Timeout   int64  `yaml:"timeout"`
	} `yaml:"base"`
	//YTDL interface{} `yaml:"youtube_dl"`
}

var cfg = Config{
	Base: struct {
		Token     string "yaml:\"tg_token\""
		Debug     bool   "yaml:\"debug\""
		SizeLimit int64  `yaml:"size_limit"`
		Timeout   int64  `yaml:"timeout"`
	}{
		Token:     "0000000000:11111111111111111111111111111111111",
		Debug:     false,
		Timeout:   30,
		SizeLimit: 100 << 20,
	},
	// TT: *downloader.NewTTParse(),
	// IG: downloader.IG{},
	// YTDL: downloader.YTdl{
	// 	Format: "18/17,bestvideo[height<=720][ext=mp4]+worstaudio,(mp4)[ext=mp4][vcodec^=h26],worst[width>=480][ext=mp4],worst[ext=mp4]",
	// },
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
	//cfg.IG.SplashRequest = base64.StdEncoding.EncodeToString([]byte(cfg.IG.SplashRequest))
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
	//cfg.TT = *downloader.NewTTParse()

	// ti, err := base64.StdEncoding.DecodeString(cfg.IG.SplashRequest)
	// if err != nil {
	// 	return
	// }
	// cfg.IG.SplashRequest = string(ti)
	return
}
