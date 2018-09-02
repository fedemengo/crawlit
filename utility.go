package crawlit

import (
	"fmt"
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

// ValidURL check if a URL is valid given a crawling configuration
func ValidURL(config CrawlConfig, id int, elem urlData, nextURL *url.URL) (valid bool) {
	valid = true

	if config.MaxDistance == -1 {
		if elem.url.Host != config.seedURLs[id].Host {
			valid = false
		}
	} else if elem.dist+1 > config.MaxDistance {
		valid = false
	}

	if config.Restrict && nextURL.Host != config.seedURLs[id].Host {
		valid = false
	}

	ClearURL(nextURL)
	return
}

// GetURL is a custom wrapper around client.Get for better handling response status
func GetURL(c *http.Client, url string) (res *http.Response, err error) {
	res, err = c.Get(url)
	if err != nil {
		if netError, ok := err.(net.Error); ok && netError.Timeout() {
			err = fmt.Errorf("Timeout on \"%s\" skipping", url)
		} else {
			err = fmt.Errorf("Error on \"%s\" skipping", url)
		}
	} else if code := res.StatusCode; code != 200 {
		err = fmt.Errorf("StatusCode %d on \"%s\" skipping", code, url)
	}
	return
}
