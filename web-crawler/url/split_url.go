package url

import (
	"errors"
	"fmt"
	"strings"
)

// SplitURL split the given URL in hostname and pathname
func SplitURL(url string) (host string, path string, err error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			fmt.Println(url)
			fmt.Println(host + ":" + path)
			return
		}
	}()

	if strings.Index(url, "://") == -1 {
		err = errors.New("bad url format")
		return
	}
	prot := url[:strings.Index(url, "://")] + "://"

	var hostname string
	if index := strings.Index(url[len(prot):], "/"); index == -1 {
		hostname = url[len(prot):]
	} else {
		hostname = url[len(prot) : len(prot)+index]
	}
	pathname := url[len(prot)+len(hostname):]

	host = prot + hostname
	path = pathname
	return
}
