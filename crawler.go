package crawlit

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fedemengo/go-data-structures/queue"
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

// Handler callback type
type Handler func(res *http.Response) error

type queueElem struct {
	url  *url.URL
	dist int
}

// Crawler represent an object to extrapolate link from website
type Crawler struct {
	result  chan map[string][]string
	routine int
}

// NewCrawler creates a new Crawler instance
func NewCrawler() *Crawler {
	c := new(Crawler)
	c.result = make(chan map[string][]string)
	c.routine = 0
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
	routines := 0
	// spawn a routine for each seed to crawl
	for i := range config.SeedURLs {
		collect[i] = make([]string, 0)
		go c.crawl(config, i, &collect[i], done, handler)
		routines++
	}

	// routine listen for result and termination
	go func() {
		for routines > 0 {
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

func (c *Crawler) crawl(config CrawlConfig, id int, collect *[]string, done chan int, handler Handler) {
	defer func() {
		done <- id
	}()

	startURL, err := url.Parse(config.SeedURLs[id])
	if err != nil {
		return
	}

	discovered := 0
	// keep track of the queued url
	inQueue := make(map[string]bool)
	// keep track of the crawled url (res.StatusCode == 200)
	crawled := make(map[string]bool)

	// one http.Client for each routine
	client := http.Client{
		Timeout: time.Duration(time.Duration(config.Timeout) * time.Second),
	}

	q := queue.NewQueue()
	q.Push(queueElem{url: startURL, dist: 0})
	inQueue[startURL.String()] = true

	for q.Size() > 0 {
		elem := q.Pop().(queueElem)

		res, err := GetURL(&client, elem.url)
		if err != nil {
			continue
		}

		cleanURL := ClearURL(res.Request.URL)
		if _, ok := crawled[cleanURL]; ok {
			continue
		}

		discovered++
		if discovered > config.MaxURLs {
			return
		}

		// save new URL whose request went through
		crawled[cleanURL] = true
		elem.url, _ = url.Parse(cleanURL)
		*collect = append(*collect, elem.url.String())
		if err = handler(res); err != nil {
			return
		}

		body := res.Body
		defer body.Close()
		doc, err := goquery.NewDocumentFromReader(body)
		if err != nil {
			//fmt.Println("ERROR: can't read body")
			continue
		}

		selector := doc.Find("a")
		for i := range selector.Nodes {

			href, _ := selector.Eq(i).Attr("href")
			nextURL, err := elem.url.Parse(href)
			if err != nil {
				continue
			}

			if !ValidURL(config, id, elem, startURL, nextURL) {
				continue
			}

			nextElem := queueElem{url: nextURL, dist: elem.dist + 1}
			if _, ok := inQueue[nextElem.url.String()]; ok {
				continue
			}

			// avoid duplicate URL in queue
			q.Push(nextElem)
			inQueue[nextElem.url.String()] = true
		}
	}
}
