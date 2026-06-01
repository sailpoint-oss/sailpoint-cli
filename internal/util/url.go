package util

import (
	"net/url"
	"path"

	"github.com/charmbracelet/log"
)

func ResourceUrl(endpoint string, resourceParts ...string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Error("invalid endpoint", "endpoint", endpoint, "error", err)
		return ""
	}
	u.Path = path.Join(append([]string{u.Path}, resourceParts...)...)
	return u.String()
}
