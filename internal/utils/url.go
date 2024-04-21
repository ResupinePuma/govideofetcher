package utils

import (
	"errors"
	"regexp"
	"strings"

	nurl "net/url"
)

var (
	ErrInvalidUrl = errors.New("Invalid url")
)

func ExtractUrlAndText(str string) (u *nurl.URL, label string, err error) {
	re := regexp.MustCompile(`(?m)(https?:\/\/[^\s]+)`)
	url := re.FindString(str)
	if url == "" {
		err = errors.New("invalid url")
		return
	}
	u, errp := nurl.ParseRequestURI(url)
	if errp != nil {
		err = errors.New("invalid url")
		return
	}
	label = strings.TrimSpace(strings.Replace(str, u.String(), "", -1))
	return
}
