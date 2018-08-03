package main

import (
	"fmt"
	"github.com/fedemengo/search-engine/web-crawler/crawler"
	"os"
)

func main() {
	seedURLs := os.Args[1:]

	c := crawler.NewCrawler(seedURLs, false, 1, 3)
	c.Crawl()

	//fmt.Println("SOME OTHER STUFF")
	//for x := 0; x < 30000000000; x++ {
	//	if x%1000000000 == 0 {
	//		fmt.Println(x)
	//	}
	//}

	foundURLs := c.Result()
	for seedURL, foundURLs := range foundURLs {
		fmt.Println(len(foundURLs), "found for", seedURL)
		for _, url := range foundURLs {
			fmt.Println(" - " + url)
		}
	}
}
