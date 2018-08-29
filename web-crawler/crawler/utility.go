package crawler

import (
	"net/url"
	"regexp"
	"fmt"
	"net"
	"net/http"
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

func LogResponse(url string, res *http.Response, err error) (skip bool) {
	skip = false
	if err != nil {
		if netError, ok := err.(net.Error); ok && netError.Timeout() {
			fmt.Println("ERROR TIMEOUT: on", "\""+url+"\"")
		} else {
			fmt.Println("ERROR: can't crawl", "\""+url+"\"")
		}
		skip = true
	} else if res.StatusCode == 404 {
		fmt.Println("ERROR 404: skipping", "\""+url+"\"")
		skip = true
	}
	return
}
