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
	SeedURLs    []string
	MaxURLs     int
	MaxDistance int
	Timeout     int
	Restrict    bool
	seedURLs    []*url.URL
}

// Handler callback type
type Handler func(res *http.Response) error

type urlData struct {
	url  *url.URL
	dist int
}

// Crawler represent an object to extrapolate link from website
type Crawler struct {
	result  chan map[int][]string
	routine int
}

// NewCrawler creates a new Crawler instance
func NewCrawler() *Crawler {
	c := new(Crawler)
	c.result = make(chan map[int][]string)
	c.routine = 0
	return c
}

// Result will return the result of the crawling, blocking
func (c *Crawler) Result() (urls map[int][]string) {

	urls = <-c.result
	return
}

// Crawl is the public method used to start the crawling
func (c *Crawler) Crawl(config CrawlConfig, handler Handler) {

	results := make(map[int][]string)
	// notify when routines are done
	quit := make(chan int)

	config.seedURLs = make([]*url.URL, len(config.SeedURLs))
	for i, u := range config.SeedURLs {
		url, err := url.Parse(u)
		url = url.ResolveReference(url)
		if err == nil {
			config.seedURLs[i] = url
		}
	}

	collect := make([][]string, len(config.SeedURLs))
	// spawn a routine for each seed to crawl
	for i := range config.seedURLs {
		collect[i] = make([]string, 0)
		go c.crawl(config, i, &collect[i], quit, handler)
	}

	// routine listen for result and termination
	go func() {
		for seeds := len(config.seedURLs); seeds > 0; {
			select {
			// listen for completed seed
			case id := <-quit:
				seeds--
				results[id] = collect[id]
				fmt.Println("COMPLETE", config.SeedURLs[id])
			}
		}
		close(quit)
		c.result <- results
	}()
}

func (c *Crawler) isValid(config CrawlConfig, id int, elem urlData, nextURL *url.URL) (uData urlData, valid bool) {
	uData = urlData{url: nextURL, dist: elem.dist + 1}
	valid = true

	if config.MaxDistance == -1 {
		if elem.url.Host != config.seedURLs[id].Host {
			valid = false
		}
	} else if uData.dist > config.MaxDistance {
		valid = false
	}

	if config.Restrict && nextURL.Host != config.seedURLs[id].Host {
		valid = false
	}

	ClearURL(uData.url)
	return
}

func (c *Crawler) crawl(config CrawlConfig, id int, collect *[]string, quit chan int, handler Handler) {

	// initialize a client for each routine
	client := http.Client{
		Timeout: time.Duration(time.Duration(config.Timeout) * time.Second),
	}

	discovered := 0
	// keep track of the queued url
	inQueue := make(map[string]bool)
	// keep track of the crawled url (res.StatusCode == 200)
	crawled := make(map[string]bool)
	q := queue.NewQueue()

	// push the seed in queue
	q.Push(urlData{url: config.seedURLs[id], dist: -1})
	inQueue[config.seedURLs[id].String()] = true

	for q.Size() > 0 {
		elem := q.Pop().(urlData)

		cleanURL := elem.url.String()
		res, err := GetURL(&client, cleanURL)
		if err != nil {
			continue
		}

		cleanURL = ClearURL(res.Request.URL)
		if _, ok := crawled[cleanURL]; ok {
			continue
		}

		discovered++
		if discovered > config.MaxURLs {
			quit <- id
			return
		}

		// save new URL whose request went through
		crawled[cleanURL] = true
		res.Request.URL, _ = url.Parse(cleanURL)
		elem.url = res.Request.URL
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

			nextElem, valid := c.isValid(config, id, elem, nextURL)
			if !valid {
				continue
			}

			if _, ok := inQueue[nextElem.url.String()]; ok {
				continue
			}

			q.Push(nextElem)
			// avoid duplicate URL in queue
			inQueue[nextElem.url.String()] = true

		}
	}
	quit <- id
}
