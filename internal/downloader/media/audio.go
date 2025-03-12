package media

import (
	"io"
	"os"
)

type Audio struct {
	Title     string
	URL       string
	Dir       string
	Thumbnail io.Reader
	Duration  float64
	Artist    string
	Reader    io.ReadCloser
}

func NewAudio(title, url string, r io.ReadCloser) *Audio {
	v := Audio{
		Title:  title,
		URL:    url,
		Reader: r,
	}
	if v.Title == "" {
		v.Title = url
	}
	return &v
}

func (v *Audio) Close() error {
	if v.Reader == nil {
		return nil
	}
	os.RemoveAll(v.Dir)
	return v.Reader.Close()
}

func (v *Audio) NeedsUpload() bool {
	return true
}

// UploadData gets the file name and an `io.Reader` for the file to be uploaded. This
// must only be called when the file needs to be uploaded.
func (v *Audio) UploadData() (string, io.Reader, error) {
	return v.Title, v.Reader, nil
}

// SendData gets the file data to send when a file does not need to be uploaded. This
// must only be called when the file does not need to be uploaded.
func (v *Audio) SendData() string {
	return v.Title
}
