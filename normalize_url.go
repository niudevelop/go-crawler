package main

import (
	"fmt"
	"net/url"
)

func normalizeURL(URL string) (string, error) {
	parsed, err := url.Parse(URL)
	if err != nil {
		return "", err
	}
	normalized := fmt.Sprintf("%s%s", parsed.Host, parsed.Path)
	return normalized, nil
}
