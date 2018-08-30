package crawler

import (
	"net"
	"net/http"
	"net/url"
)

// ClearURL sanitize a URL by removing unnecessary field
func ClearURL(u *url.URL) string {
	u.Fragment = ""
	u.RawQuery = ""
	u = u.ResolveReference(u)

	plainURL := u.String()
	if uLen := len(plainURL); uLen > 0 && plainURL[uLen-1] == '/' {
		plainURL = plainURL[:uLen-1]
	}
	return plainURL
}

// LogResponse print the status of the response
func LogResponse(url string, res *http.Response, err error) (skip bool) {
	skip = false
	if err != nil {
		if netError, ok := err.(net.Error); ok && netError.Timeout() {
			//fmt.Printf("ERROR TIMEOUT: on \"%s\"\n", url)
		} else {
			//fmt.Printf("ERROR: can't crawl \"%s\"\n", url)
		}
		skip = true
	} else if code := res.StatusCode; code != 200 {
		//fmt.Printf("ERROR %d: skipping \"%s\"\n", code, url)
		skip = true
	}
	return
}
