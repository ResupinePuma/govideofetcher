package options

import "net/http"

type DownloaderOpts struct {
	SizeLimit int64 `yaml:"size_limit"`
	Timeout   int64 `yaml:"timeout"`
	AdminID   int64 `yaml:"admin_id"`

	YTDL YTDLOptions `yaml:"youtube_dl"`
}

type YTDLOptions struct {
	Format     string `yaml:"format"`
	Executable string `yaml:"executable"`
	FFmpeg     string `yaml:"ffmpeg"`
	Headers    http.Header
}

type RedditOptions struct {
	ID     string `yaml:"id"`
	Secret string `yaml:"secret"`
}
