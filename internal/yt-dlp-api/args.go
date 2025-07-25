package ytdlpapi

import "net/http"

type BodyArgs struct {
	URL                 string      `json:"url,omitempty"`
	Format              string      `json:"format,omitempty"`
	Extension           string      `json:"extension,omitempty"`
	MaxFilesize         string      `json:"max_filesize,omitempty"`
	ProxyURL            string      `json:"proxy_url,omitempty"`
	BufferSize          string      `json:"buffer_size,omitempty"`
	DownloadInfo        bool        `json:"write_info_json,omitempty"`
	DownloadThumb       bool        `json:"write_thumbnail,omitempty"`
	FFMpegDurationLimit *int64      `json:"ffmpeg_dur_limit,omitempty"`
	FFMpeg              string      `json:"ffmpeg,omitempty"`
	Headers             http.Header `json:"headers,omitempty"`

	MaxSize int64
}
