package util

import (
	"log"
	"net/url"
	"path"
)

func ResourceUrl(endpoint string, resourceParts ...string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Fatalf("invalid endpoint: %s (%q)", err, endpoint)
	}
	u.Path = path.Join(append([]string{u.Path}, resourceParts...)...)
	return u.String()
}
