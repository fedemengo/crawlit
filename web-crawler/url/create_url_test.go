package url

import "testing"

func testURL(t *testing.T, exp string, res string) {
	if exp != res {
		t.Error("Expected", exp, "get", res)
	}
}

func TestCreateURL(t *testing.T) {
	host := "https://www.example.org"
	path := "/some/path"
	testURL(t, "https://www.example.org/some/path", CreateURL(host, path)[0])

	host = "https://www.example.org"
	path = "//example2.me/else"
	testURL(t, "https://example2.me/else", CreateURL(host, path)[0])

	host = "https://www.example.org"
	path = "/something/else?x=y"
	testURL(t, "https://www.example.org/something/else", CreateURL(host, path)[0])

	host = "https://www.example.org"
	path = "/something/else#x=y/what"
	testURL(t, "https://www.example.org/something/else", CreateURL(host, path)[0])

	host = "https://www.example.org"
	path = "/something/else%adad%ada"
	testURL(t, "https://www.example.org/something/else", CreateURL(host, path)[0])
}
