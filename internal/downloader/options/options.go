package options

import "net/http"

type DownloaderOpts struct {
	SizeLimit       int64  `yaml:"size_limit"`
	Timeout         int64  `yaml:"timeout"`
	AdminID         int64  `yaml:"admin_id"`
	DownloaderProxy string `yaml:"dn_proxy"`

	YTDL YTDLOptions `yaml:"youtube_dl"`
}

type YTDLOptions struct {
	Format     string `yaml:"format"`
	Executable string `yaml:"executable"`
	FFmpeg     string `yaml:"ffmpeg"`
	APIAddr    string `yaml:"api_addr"`
	Proxies    string
	Headers    http.Header

	isSet bool
}

func (o *YTDLOptions) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type t YTDLOptions
	tmp := t{
		isSet: true,
	}

	err := unmarshal(&tmp)
	if err != nil {
		return err
	}

	*o = YTDLOptions(tmp)
	return nil
}

func (o *YTDLOptions) IsSet() bool {
	return o.isSet
}
