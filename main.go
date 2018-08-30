package main

import (
	"fmt"
	"os"
	"net/http"

	"github.com/fedemengo/search-engine/web-crawler"
)

func main() {
	seedURLs := os.Args[1:]

	/**
	 * urls
	 * restricted
	 * distance
	 * timeout
	 * max number of url
	 */
	c := crawler.NewCrawler(seedURLs, false, 1, 5, 500)
	c.Crawl(func(res *http.Response) (err error) {
		/*		for k, v := range res.Header {
					fmt.Println(k)
					for _, d := range v {
						fmt.Print(d, " ")
					}
					fmt.Println()
				}
		*/
		fmt.Println(" -> " + res.Request.URL.String())
		return nil
	})

	//fmt.Println("SOME OTHER STUFF")
	//for x := 0; x < 30000000000; x++ {
	//	if x%1000000000 == 0 {
	//		fmt.Println(x)
	//	}
	//}

	foundURLs := c.Result()
	for i, foundURLs := range foundURLs {
		fmt.Println(len(foundURLs), "found for", seedURLs[i])
		for _, url := range foundURLs {
			fmt.Println(" - " + url)
		}
	}
}
