package crawlit

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

func TestCrawler(t *testing.T) {
	// pass argument with `go test -args arg1 arg2 ...`

	c := NewCrawler()

	config := CrawlConfig{
		SeedURLs:    os.Args[1:2],
		MaxURLs:     20,
		MaxDistance: 0,
		Timeout:     3,
		Restrict:    false,
	}

	c.Crawl(config, func(res *http.Response) (err error) {
		fmt.Println(" >> " + res.Request.URL.String())
		//return crawlit.StopCrawl
		return nil
	})

	config = CrawlConfig{
		SeedURLs:    os.Args[2:],
		MaxURLs:     10,
		MaxDistance: 1,
		Timeout:     3,
		Restrict:    true,
	}

	c.Crawl(config, func(res *http.Response) (err error) {
		fmt.Println(" -- " + res.Request.URL.String())
		return SkipURL
	})

	fmt.Println("SOME OTHER STUFF")

	// Consume one crawling
	c.Result()

	// Consume another crawling
	foundURLs := c.Result()
	for seed, urls := range foundURLs {
		fmt.Println(len(urls), "found for", seed)
		for _, u := range urls {
			fmt.Println(" - " + u)
		}
	}

}
