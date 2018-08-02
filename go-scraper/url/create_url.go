package url

import (
	"fmt"
	"regexp"
	"strings"
)

func removeTerminal(url string) string {
	mail := regexp.MustCompile(`mailto.*`)
	slash := regexp.MustCompile(`/$`)
	return slash.ReplaceAllString(mail.ReplaceAllString(url, ""), "")
}

// CreateURL creates a sanityzed URL from the given host and pathname
func CreateURL(h string, p string) (urls []string) {
	urls = []string{""}

	if h == "" {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("RECOVER", h, p)
		}
	}()

	special := regexp.MustCompile(`([?&#%].*)`)
	p = special.ReplaceAllString(p, "")

	urls[0] = h + p
	if strings.Index(p, "//") == 0 {
		urls[0] = h[:strings.Index(h, "//")] + p
	} else if strings.Index(p, "http") == 0 {
		urls[0] = p
	}

	if index := strings.LastIndex(urls[0], "http"); index != 0 {
		nextURL := removeTerminal(urls[0][index:])
		newURL := removeTerminal(urls[0][:index])
		h, p, _ := SplitURL(nextURL)
		urls = append(CreateURL(h, p), newURL)
	} else {
		urls[0] = removeTerminal(urls[0])
	}

	return
}
