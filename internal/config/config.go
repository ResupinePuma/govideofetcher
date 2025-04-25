package config

import (
	"errors"
	"os"
	"syscall"
	"videofetcher/internal/downloader/options"

	"gopkg.in/yaml.v2"
)

var ErrCfg = errors.New("error reading config.yml")

type Config struct {
	Token                  string `yaml:"tg_token"`
	AdminId                int64  `yaml:"admin_id"`
	Debug                  bool   `yaml:"debug"`
	options.DownloaderOpts `yaml:"downloader"`
}

var cfg = Config{
	Token: "0000000000:11111111111111111111111111111111111",
	Debug: false,
	DownloaderOpts: options.DownloaderOpts{
		Timeout:   30,
		SizeLimit: 100 << 20,
		YTDL: options.YTDLOptions{
			Format:     "18/17,bestvideo[height<=720][ext=mp4]+worstaudio,(mp4)[ext=mp4][vcodec^=h26],worst[width>=480][ext=mp4],worst[ext=mp4]",
			Executable: "yt-dlp",
			APIAddr: "http://127.0.0.1/convert",
		},
	},
}

func NewDefaultConfig() (err error) {
	err = os.Mkdir("configs", 0644)
	if err != nil && !errors.Is(err, syscall.EEXIST) {
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
