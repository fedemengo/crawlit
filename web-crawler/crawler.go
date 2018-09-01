package crawler

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fedemengo/go-utility/data-structures/queue"
)

// Handler callback type
type Handler func(res *http.Response) error

type urlData struct {
	pURL *url.URL
	dist int
}

// Crawler represent an object to extrapolate link from website
type Crawler struct {
	seedURLs  []*url.URL
	restrict  bool
	distance  int
	timeout   int
	maxURL    int
	maxQueued int
	discURLs  [][]string
	result    chan map[int][]string
}

// NewCrawler creates a new Crawler instance
func NewCrawler(urls []string, restrict bool, distance, timeout, maxURL int) *Crawler {
	c := new(Crawler)
	for _, u := range urls {
		next, err := url.Parse(u)
		next = next.ResolveReference(next)
		if err == nil {
			c.seedURLs = append(c.seedURLs, next)
		}
	}
	c.restrict = restrict
	c.distance = distance
	c.timeout = timeout
	c.maxURL = maxURL
	c.maxQueued = 100 * c.maxURL
	c.discURLs = make([][]string, len(c.seedURLs))
	c.result = make(chan map[int][]string)
	return c
}

// Result will return the result of the crawling, blocking
func (c *Crawler) Result() (urls map[int][]string) {
	urls = <-c.result
	return
}

// Crawl is the public method used to start the crawling
func (c *Crawler) Crawl(handler Handler) {

	results := make(map[int][]string)
	// notify when routines are done
	quit := make(chan int)

	// spawn a routine for each seed to crawl
	for i := range c.seedURLs {
		go c.crawl(i, handler, quit)
	}

	// routine listen for result and termination
	go func() {
		for seeds := len(c.seedURLs); seeds > 0; {
			select {
			// listen for completed seed
			case id := <-quit:
				seeds--
				results[id] = c.discURLs[id]
				fmt.Println("COMPLETE", c.seedURLs[id].String())
			}
		}
		close(quit)
		c.result <- results
	}()
}

func (c *Crawler) isValid(id int, elem urlData, nextURL *url.URL) (uData urlData, valid bool) {
	uData = urlData{pURL: nextURL, dist: elem.dist + 1}
	valid = true

	if c.distance == -1 {
		if elem.pURL.Host != c.seedURLs[id].Host {
			valid = false
		}
	} else if uData.dist > c.distance {
		valid = false
	}

	if c.restrict && nextURL.Host != c.seedURLs[id].Host {
		valid = false
	}

	ClearURL(uData.pURL)
	return
}

func (c *Crawler) crawl(id int, handler Handler, quit chan int) {

	// initialize a client for each routine
	client := http.Client{
		Timeout: time.Duration(time.Duration(c.timeout) * time.Second),
	}

	discovered := 0
	// keep track of the queued url
	inQueue := make(map[string]bool)
	// keep track of the crawled url (res.StatusCode == 200)
	crawled := make(map[string]bool)
	q := queue.NewQueue()

	// push the seed in queue
	q.Push(urlData{pURL: c.seedURLs[id], dist: -1})
	inQueue[c.seedURLs[id].String()] = true

	for q.Size() > 0 {
		elem := q.Pop().(urlData)

		cleanURL := elem.pURL.String()
		res, err := GetURL(&client, cleanURL)
		if err != nil {
			continue
		}

		cleanURL = ClearURL(res.Request.URL)
		if _, ok := crawled[cleanURL]; ok {
			continue
		}

		discovered++
		if discovered > c.maxURL {
			return
		}

		// save new URL whose request went through
		crawled[cleanURL] = true
		res.Request.URL, _ = url.Parse(cleanURL)
		elem.pURL = res.Request.URL
		c.discURLs[id] = append(c.discURLs[id], elem.pURL.String())
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
			nextURL, err := elem.pURL.Parse(href)
			if err != nil {
				continue
			}

			nextElem, valid := c.isValid(id, elem, nextURL)
			if !valid {
				continue
			}

			if _, ok := inQueue[nextElem.pURL.String()]; ok {
				continue
			}

			if q.Size() < c.maxQueued {
				q.Push(nextElem)
				// avoid duplicate URL in queue
				inQueue[nextElem.pURL.String()] = true
			}
		}
	}
	quit <- id
}
