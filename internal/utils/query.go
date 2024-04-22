package utils

import "net/url"

func GenerateQuery(args map[string]string) string {
	data := url.Values{}
	for k, v := range args {
		data.Set(k, v)
	}
	return data.Encode()

}
