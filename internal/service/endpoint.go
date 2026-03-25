package service

import (
	"net/url"
	"strings"
	"zpcli/internal/logx"
)

func BuildEndpointURL(domainURL string) string {
	logger := logx.Logger("service.endpoint")
	original := domainURL
	if !strings.HasPrefix(domainURL, "http://") && !strings.HasPrefix(domainURL, "https://") {
		domainURL = "http://" + domainURL
	}
	u, err := url.Parse(domainURL)
	if err != nil {
		logger.Warn("build endpoint parse failed", "input", original, "current", domainURL, "error", err)
		return domainURL
	}
	if u.Path == "" || u.Path == "/" {
		if !strings.HasSuffix(domainURL, "/") {
			domainURL += "/"
		}
		domainURL += "api.php/provide/vod"
	}
	logger.Debug("build endpoint url", "input", original, "output", domainURL)
	return domainURL
}
