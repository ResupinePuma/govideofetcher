package video

import (
	"io"
	"os"
)

type Video struct {
	Title  string
	URL    string
	Dir    string
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
	os.Remove(v.Dir)
	return v.Reader.Close()
}

func (v *Video) NeedsUpload() bool {
	return true
}

// UploadData gets the file name and an `io.Reader` for the file to be uploaded. This
// must only be called when the file needs to be uploaded.
func (v *Video) UploadData() (string, io.Reader, error) {
	return v.Title, v.Reader, nil
}

// SendData gets the file data to send when a file does not need to be uploaded. This
// must only be called when the file does not need to be uploaded.
func (v *Video) SendData() string {
	return v.Title
}
