package parser

import (
	"errors"

	"golang.org/x/net/html"
)

// Parse extract the property from the given tag
func Parse(t html.Token, tag string, prop string) (value string, err error) {
	if t.Data != tag {
		err = errors.New("Tag doesn't match")
		return
	}
	for _, attr := range t.Attr {
		if attr.Key == prop {
			value = attr.Val
		}
	}
	return
}
