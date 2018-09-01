package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/fedemengo/search-engine/web-crawler"
)

func main() {
	seedURLs := os.Args[1:]

	// urls
	// restricted		keep only URL from seed host
	// distance			-1 scrape all host
	// 					N scrape up to dist N from url
	// timeout
	// max number of url
	c := crawler.NewCrawler(seedURLs, false, -1, 3, 500)
	c.Crawl(func(res *http.Response) (err error) {
		fmt.Println(" -> " + res.Request.URL.String())
		return nil
	})

	// fmt.Println("SOME OTHER STUFF")

	foundURLs := c.Result()
	for i, foundURLs := range foundURLs {
		fmt.Println(len(foundURLs), "found for", seedURLs[i])
		for _, url := range foundURLs {
			fmt.Println(" - " + url)
		}
	}
}
