## Web crawler

Stateless web crawler configurable on each crawling call. All host is crawled concurrently, so `c.Crawl(config, handler)` is **non-blocking** and it calls the `handler` function for every valid url discovered. 

The method `c.Result()`, on the other hand, is blocking and once called it consume the result of **one** crawling. At the moment the order of crawling collection (the call to `c.Result()`) doesn't reflect the order in which the crawling started.

## Configuration

```go
type CrawlConfig struct {
	SeedURLs    []string
	MaxURLs     int         // maximum number of URL to crawl
	MaxDistance int         // maximum page distance to crawl: -1 for infinite, 0 for crawling the whole host
	Timeout     int         // maximum timeout to wait for response
	Restrict    bool        // restrict crawling to seed host
}
```

## Example

```go
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
```