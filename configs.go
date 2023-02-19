package main

import (
	"errors"
	"videofetcher/downloader"
	"os"

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
	YTDL downloader.YTdl   `yaml:"youtube_dl"`
}

func NewDefaultConfig() (err error) {
	cfg := Config{
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
			SplashURL: "http://127.0.0.1:8050/execute",
			SplashRequest: `function main(splash, args)
			splash:on_request(function(request)
				request.headers['user-agent'] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36"
			end)
			splash:go(args.url)
			assert(splash:wait(0.5))
			splash:runjs('document.getElementsByName("url")[0].value="%s";document.getElementsByTagName("form")[0].submit();')
			assert(splash:wait(3)) 
			local vurl = splash:evaljs('document.getElementsByClassName("abuttons mb-0")[0].children[0].href')
			local title = splash:evaljs('document.getElementsByClassName("videotikmate-middle center")[0].children[1].innerText')                       
			return {{
				title = title,
				vurl = vurl
			}}
			end
			`,
			TTUrl: "https://tikmate.online/?lang=nl",
		},
		YTDL: downloader.YTdl{
			Format:     "18/17,bestvideo[height<=480][ext=mp4]+worstaudio,(mp4)[ext=mp4][vcodec^=h26],worst[ext=mp4]",
		},
	}
	err = os.Mkdir("configs", 0644)
	if err != nil {
		return
	}
	f, err := os.Create("configs/config.yml")
	if err != nil {
		return
	}
	defer f.Close()

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
	return
}
