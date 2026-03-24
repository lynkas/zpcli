package service

import (
	"net/url"
	"strings"
)

func BuildEndpointURL(domainURL string) string {
	if !strings.HasPrefix(domainURL, "http://") && !strings.HasPrefix(domainURL, "https://") {
		domainURL = "http://" + domainURL
	}
	u, err := url.Parse(domainURL)
	if err != nil {
		return domainURL
	}
	if u.Path == "" || u.Path == "/" {
		if !strings.HasSuffix(domainURL, "/") {
			domainURL += "/"
		}
		domainURL += "api.php/provide/vod"
	}
	return domainURL
}
