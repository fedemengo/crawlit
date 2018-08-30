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
	URLs      []*url.URL
	Restrict  bool
	Distance  int
	timeout   int
	maxURL    int
	maxQueued int
	resultCh  chan map[int][]string
}

// NewCrawler creates a new Crawler instance
func NewCrawler(urls []string, restrict bool, distance, timeout, maxURL int) *Crawler {
	c := new(Crawler)
	for _, u := range urls {
		next, err := url.Parse(u)
		if err == nil {
			c.URLs = append(c.URLs, next)
		}
	}
	c.Restrict = restrict
	c.Distance = distance
	c.timeout = timeout
	c.maxURL = maxURL
	c.maxQueued = 100 * c.maxURL
	c.resultCh = make(chan map[int][]string)
	return c
}

// Result will return the result of the crawling, blocking
func (c *Crawler) Result() (urls map[int][]string) {
	urls = <-c.resultCh
	return
}

// Crawl is the public method used to start the crawling
func (c *Crawler) Crawl(handler Handler) {
	// store urls retrieved from any given seed
	result := make(map[int][]string)
	// channels for each seed
	chURLs := make([]chan *url.URL, len(c.URLs))
	for i := range chURLs {
		chURLs[i] = make(chan *url.URL)
	}
	// notify when routines are done
	quit := make(chan int)

	// spawn a routine for each seed to crawl
	for i := range c.URLs {
		go c.crawl(c.URLs[i], chURLs[i], handler)
	}

	// spawn a routine to listen on every seed channel
	for i, ch := range chURLs {
		go func(c chan *url.URL, id int) {
			for u := range c {
				// save the results
				result[id] = append(result[id], u.String())
			}
			quit <- id
		}(ch, i)
	}

	// routine listen for result and termination
	go func() {
		for seeds := len(c.URLs); seeds > 0; {
			select {
			// listen for completed seed
			case id := <-quit:
				seeds--
				fmt.Println("COMPLETE", c.URLs[id].String())
			}
		}
		close(quit)
		c.resultCh <- result
		close(c.resultCh)
	}()
}

func (c *Crawler) isValid(seedURL *url.URL, elem urlData, nextURL *url.URL) (uData urlData, valid bool) {
	uData = urlData{pURL: nextURL, dist: elem.dist + 1}
	valid = true

	if c.Distance == -1 {
		if elem.pURL.Host != seedURL.Host {
			valid = false
			return
		}
	} else if uData.dist > c.Distance {
		valid = false
		return
	}

	if c.Restrict && nextURL.Host != seedURL.Host {
		valid = false
		return
	}

	ClearURL(uData.pURL)
	return
}

func (c *Crawler) crawl(seedURL *url.URL, chURL chan *url.URL, handler Handler) {
	// one completed close channel
	defer func() {
		close(chURL)
	}()

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

	// resolve references if present
	seedURL = seedURL.ResolveReference(seedURL)
	// push the seed in queue
	q.Push(urlData{pURL: seedURL, dist: -1})
	inQueue[seedURL.String()] = true

	for q.Size() > 0 {
		elem := q.Pop().(urlData)

		plainURL := elem.pURL.String()
		res, err := client.Get(plainURL)
		if skip := LogResponse(plainURL, res, err); skip {
			continue
		}

		reqURL := ClearURL(res.Request.URL)
		if _, ok := crawled[reqURL]; ok {
			continue
		}

		discovered++
		if discovered > c.maxURL {
			return
		}

		// save new URL whose request went through
		crawled[reqURL] = true
		res.Request.URL, _ = url.Parse(reqURL)
		elem.pURL = res.Request.URL
		chURL <- elem.pURL
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

			nextElem, valid := c.isValid(seedURL, elem, nextURL)
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
}
