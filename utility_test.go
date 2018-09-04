package crawlit

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestClearURL(t *testing.T) {
	stringURL := "https://www.google.com/search?source=hp&ei=U0SOW7PgNYfmswHpt6KgCA&q=foo&oq=foo&gs_l=psy-ab.3..0l10.6392.6731.0.6925.4.3.0.0.0.0.140.275.0j2.2.0....0...1.1.64.psy-ab..2.2.275.0...0.FCYJ3WfrcQc"
	url, _ := url.Parse(stringURL)

	stringURL = ClearURL(url)
	exp := "https://www.google.com/search"

	if stringURL != exp {
		t.Error("Expected", exp, "got", stringURL)
	}

	expURL, _ := url.Parse(exp)
	if url.String() != expURL.String() {
		t.Error("Expected", expURL, "got", url)
	}
}

func TestValidURL(t *testing.T) {

	config := CrawlConfig{
		SeedURLs:    []string{"https://www.google.com/"},
		MaxURLs:     10,
		MaxDistance: 2,
		Timeout:     3,
		Restrict:    false,
	}

	startURL, _ := url.Parse(config.SeedURLs[0])
	url, _ := url.Parse(config.SeedURLs[0] + "/search")
	nextURL, _ := url.Parse(config.SeedURLs[0] + "/search/something")
	qElem := queueElem{
		url:  url,
		dist: 1,
	}

	valid := ValidURL(config, qElem, startURL, nextURL)
	if !valid {
		t.Error("Expected", nextURL, "to be valid")
	}

	config.MaxDistance = 1
	valid = ValidURL(config, qElem, startURL, nextURL)
	if valid {
		t.Error("Expected", nextURL, "to NOT be valid")
	}

	config.Restrict = true
	nextURL, _ = url.Parse("https://www.goggle.com/search/something")
	valid = ValidURL(config, qElem, startURL, nextURL)
	if valid {
		t.Error("Expected", nextURL, "to NOT be valid")
	}
}

func TestGetURL(t *testing.T) {
	client := http.Client{
		Timeout: time.Duration(time.Duration(3) * time.Second),
	}

	fileURL, _ := url.Parse("http://www.tandfonline.com/")
	_, err := GetURL(&client, fileURL)
	if err == nil {
		t.Error("Expected timeout error")
	}

	fileURL, _ = url.Parse("http://www.goggle.com/asd")
	_, err = GetURL(&client, fileURL)
	if err == nil {
		t.Error("Expected error 404")
	}
}
