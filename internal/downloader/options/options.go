package options

type DownloaderOpts struct {
	SizeLimit int64 `yaml:"size_limit"`
	Timeout   int64 `yaml:"timeout"`

	YTDL YTDLOptions `yaml:"youtube_dl"`
}

type YTDLOptions struct {
	Format     string `yaml:"format"`
	Executable string `yaml:"executable"`
}
