package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/fedemengo/crawlit"
)

func main() {

	c := crawlit.NewCrawler()

	config := crawlit.CrawlConfig{
		SeedURLs:    os.Args[1:2],
		MaxURLs:     20,
		MaxDistance: 0,
		Timeout:     3,
		Restrict:    false,
	}

	c.Crawl(config, func(res *http.Response) (err error) {
		fmt.Println(" >> " + res.Request.URL.String())
		return nil
	})

	config = crawlit.CrawlConfig{
		SeedURLs:    os.Args[2:],
		MaxURLs:     10,
		MaxDistance: 1,
		Timeout:     3,
		Restrict:    true,
	}

	c.Crawl(config, func(res *http.Response) (err error) {
		fmt.Println(" -- " + res.Request.URL.String())
		return nil
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
