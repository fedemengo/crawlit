package crawlit

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// CrawlConfig specify some parameters for the crawling
type CrawlConfig struct {
	SeedURLs []string
	// maximum number of URL to crawl
	MaxURLs int
	// maximum page distance to crawl: -1 for infinite, 0 for crawling the whole host
	MaxDistance int
	// maximum timeout to wait for response
	Timeout int
	// restrict crawling to seed host
	Restrict bool
}

// CrawlitResponse represent the data type returned from each request. Can be extended
type CrawlitResponse struct {
	URL  string
	Body *goquery.Document
}

// Handler callback type
type Handler func(res CrawlitResponse) error

// SkipURL type for handler
var SkipURL = errors.New("skip this URL")

// StopCrawl type for handler
var StopCrawl = errors.New("stop crawling")

type queueElem struct {
	url  *url.URL
	dist int
}

// Crawler represent an object to extrapolate link from website
type Crawler struct {
	result chan map[string][]string
}

// NewCrawler creates a new Crawler instance
func NewCrawler() *Crawler {
	c := new(Crawler)
	c.result = make(chan map[string][]string)
	return c
}

// Result will return the result of the crawling, blocking
func (c *Crawler) Result() (urls map[string][]string) {
	urls = <-c.result
	return
}

// Crawl is the public method used to start the crawling
func (c *Crawler) Crawl(config CrawlConfig, handler Handler) {

	results := make(map[string][]string)
	// notify when routines are done
	done := make(chan int)

	collect := make([][]string, len(config.SeedURLs))
	// spawn a routine for each seed to crawl
	for i := range config.SeedURLs {
		collect[i] = make([]string, 0)
		go crawlURL(config, i, &collect[i], done, handler)
	}

	// routine listen for result and termination
	go func() {
		for routines := len(config.SeedURLs); routines > 0; {
			select {
			// listen for completed seed
			case id := <-done:
				routines--
				results[config.SeedURLs[id]] = collect[id]
				fmt.Println("COMPLETE", config.SeedURLs[id])
			}
		}
		close(done)
		c.result <- results
	}()
}
