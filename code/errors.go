package main

var _ERRORS map[string]string = map[string]string{
	"invalid url":               "Please send valid url address",
	"size limit reached":        "Video is too large. Try another smaller video",
	"context deadline exceeded": "Timeout. Try again later",
}

func GetErrMsg(err error) string {
	if e, ok := _ERRORS[err.Error()]; ok {
		return e
	}
	return "Something went wrong. Try again or try another video"
}
