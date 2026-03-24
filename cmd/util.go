package cmd

import "zpcli/internal/service"

func buildEndpointURL(domainURL string) string {
	return service.BuildEndpointURL(domainURL)
}
