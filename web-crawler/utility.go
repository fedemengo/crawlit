package crawler

import (
	"net/url"
	"regexp"
	"fmt"
	"net"
	"net/http"
)

func ClearUrl(u *url.URL) string {
	u.Fragment = ""
	u.RawQuery = ""
	u = u.ResolveReference(u)

	special := regexp.MustCompile(`/$`)
	return special.ReplaceAllString(u.String(), "")
}

func LogResponse(url string, res *http.Response, err error) (skip bool) {
	skip = false
	if err != nil {
		if netError, ok := err.(net.Error); ok && netError.Timeout() {
			fmt.Printf("ERROR TIMEOUT: on \"%s\"\n", url)
		} else {
			fmt.Printf("ERROR: can't crawl \"%s\"\n", url)
		}
		skip = true
	} else if code := res.StatusCode; code != 200 {
		fmt.Printf("ERROR %d: skipping \"%s\"\n", code, url)
		skip = true
	}
	return
}
