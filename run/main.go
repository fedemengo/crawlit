package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/fedemengo/crawlit"
)

func main() {

	fmt.Println(os.Args[1:2])
	fmt.Println(os.Args[2:])

	c := crawlit.NewCrawler()

	config := crawlit.CrawlConfig{
		SeedURLs:    os.Args[1:2],
		MaxURLs:     10,
		MaxDistance: -1,
		Timeout:     3,
		Restrict:    false,
	}

	c.Crawl(config, func(res *http.Response) (err error) {
		fmt.Println(" -> " + res.Request.URL.String())
		return nil
	})

	/*
		foundURLs := c.Result()
		for i, foundURLs := range foundURLs {
			fmt.Println(len(foundURLs), "found for", os.Args[1:2][i])
			for _, url := range foundURLs {
				fmt.Println(" - " + url)
			}
		}
	*/

	config = crawlit.CrawlConfig{
		SeedURLs:    os.Args[2:],
		MaxURLs:     5,
		MaxDistance: -1,
		Timeout:     3,
		Restrict:    false,
	}

	c.Crawl(config, func(res *http.Response) (err error) {
		fmt.Println(" -> " + res.Request.URL.String())
		return nil
	})

	fmt.Println("SOME OTHER STUFF")

	c.Result()
	c.Result()
	/*
		foundURLs = c.Result()
		for i, foundURLs := range foundURLs {
			fmt.Println(len(foundURLs), "found for", os.Args[2:3][i])
			for _, url := range foundURLs {
				fmt.Println(" - " + url)
			}
		}
	*/
}
