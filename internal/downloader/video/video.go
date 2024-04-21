package video

import "io"

type Video struct {
	Title  string
	URL    string
	Reader io.ReadCloser
}

func NewVideo(title, url string, r io.ReadCloser) *Video {
	v := Video{
		Title:  title,
		URL:    url,
		Reader: r,
	}
	if v.Title == "" {
		v.Title = url
	}
	return &v
}

func (v *Video) Close() error {
	if v.Reader == nil {
		return nil
	}
	return v.Reader.Close()
}
