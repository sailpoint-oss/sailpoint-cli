package util

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"strings"
)

const defaultTenantDomain = "identitynow.com"

func ResourceUrl(endpoint string, resourceParts ...string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Fatalf("invalid endpoint: %s (%q)", err, endpoint)
	}
	u.Path = path.Join(append([]string{u.Path}, resourceParts...)...)
	return u.String()
}

// ParseTenantInput turns a tenant name or URL into the environment name, the
// tenant URL and the API base URL.
//
// A plain name such as "acme" uses the default identitynow.com domain. A URL
// or host such as "https://acme.identitynow-demo.com/ui" keeps its domain, so
// the API URL becomes "https://acme.api.identitynow-demo.com". A pasted API
// URL ("acme.api.identitynow-demo.com") gives the same result.
func ParseTenantInput(input string) (string, string, string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", "", fmt.Errorf("tenant name or URL is empty")
	}

	if !strings.Contains(input, ".") && !strings.Contains(input, "://") {
		return input, "https://" + input + "." + defaultTenantDomain, "https://" + input + ".api." + defaultTenantDomain, nil
	}

	if !strings.Contains(input, "://") {
		input = "https://" + input
	}

	parsed, err := url.Parse(input)
	if err != nil {
		return "", "", "", fmt.Errorf("tenant URL is not valid: %v", err)
	}
	if parsed.Scheme != "https" {
		return "", "", "", fmt.Errorf("tenant URL must use HTTPS")
	}

	host := strings.ToLower(parsed.Hostname())
	labels := strings.Split(host, ".")
	if host == "" || len(labels) < 2 {
		return "", "", "", fmt.Errorf("tenant URL has no valid host")
	}
	for _, label := range labels {
		if label == "" {
			return "", "", "", fmt.Errorf("tenant URL has no valid host")
		}
	}

	if len(labels) > 2 && labels[1] == "api" {
		labels = append(labels[:1], labels[2:]...)
	}

	name := labels[0]
	domain := strings.Join(labels[1:], ".")

	return name, "https://" + name + "." + domain, "https://" + name + ".api." + domain, nil
}
