package url

import (
	"testing"
)

func testSplit(t *testing.T, host string, path string, h string, p string) {
	if host != h {
		t.Error("Expected", host, "get", h)
	}
	if path != p {
		t.Error("Expected", path, "get", p)
	}
}

func TestSplitURL(t *testing.T) {
	host := "https://www.example.org"
	path := "/some/path"
	h1, p1 := SplitURL(host + path)
	testSplit(t, host, path, h1, p1)

	host = "https://www.example.org/somethinf/else"
	path = ""
	h1, p1 = SplitURL(host + path)
	testSplit(t, "https://www.example.org", "/somethinf/else", h1, p1)
}
