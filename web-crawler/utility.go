package crawler

import (
	"net/url"
	"regexp"
)

func removeTerminal(rawURL string) string {
	special := regexp.MustCompile(`/$`)
	return special.ReplaceAllString(rawURL, "")
}

func ClearUrl(u *url.URL) string {
	u.Fragment = ""
	u = u.ResolveReference(u)

	return removeTerminal(u.String())
}
