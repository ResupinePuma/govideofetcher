package media

import (
	"io"
	"regexp"
	"strings"
)

type Media interface {
	Close() error
	NeedsUpload() bool
	// UploadData gets the file name and an `io.Reader` for the file to be uploaded. This
	// must only be called when the file needs to be uploaded.
	UploadData() (string, io.Reader, error)
	// SendData gets the file data to send when a file does not need to be uploaded. This
	// must only be called when the file does not need to be uploaded.
	SendData() string
}

var sanitizer = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)

func SanitizeFileName(name string) string {
	name = sanitizer.ReplaceAllString(name, "")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.Trim(name, " .")

	return name
}
