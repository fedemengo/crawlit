package crawler

import (
	"fmt"
	"github.com/fedemengo/go-utility/data-structures/queue"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Handler func(res *http.Response) error

type urlData struct {
	seed string
	pUrl *url.URL	
	dist int
}

// Crawler represent an object to extrapolate link from website
type Crawler struct {
	URLs     []*url.URL
	Restrict bool
	Distance int
	maxURL	 int
	maxQueued int
	client   http.Client
	resultCh chan map[string][]string
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
	c.maxURL = maxURL
	c.maxQueued = 100 * c.maxURL
	c.client = http.Client{
		Timeout: time.Duration(time.Duration(timeout) * time.Second),
	}
	c.resultCh = make(chan map[string][]string)
	return c
}

// Result will return the result of the crawling, blocking
func (c *Crawler) Result() (urls map[string][]string) {
	urls = <-c.resultCh
	return
}

// Crawl is the public method used to start the crawling
func (c *Crawler) Crawl(handler Handler) {
	result := make(map[string][]string)
	chURLs := make([]chan urlData, len(c.URLs))
	for i := range chURLs {
		chURLs[i] = make(chan urlData)
	}
	quit := make(chan int)

	// spawn a routine for each seed to crawl
	for i := range c.URLs {
		go c.crawl(c.URLs[i], chURLs[i], handler)
	}

	// spawn a routine to listen on every seed channel
	for i, ch := range chURLs {
		go func(c chan urlData, id int) {
			for data := range c {
				// directly save the result
				result[data.seed] = append(result[data.seed], data.pUrl.String())
			}
			quit <- id
		}(ch, i)
	}

	// routine listen for result and termination
	go func() {
		for seed := len(c.URLs); seed > 0; {
			select {
			// listen for completed seed
			case id := <- quit:
				seed--
				fmt.Println("COMPLETE", c.URLs[id].String())
			}
		}
		close(quit)
		c.resultCh <- result
		close(c.resultCh)
	}()
}

func (c Crawler) crawl(newURL *url.URL, chURL chan urlData, handler Handler) {
	defer func() {
		close(chURL)
	}()
	
	discoverd := 0
	inQueue := make(map[string]bool)
	crawled := make(map[string]bool)
	q := queue.NewQueue()

	newURL = newURL.ResolveReference(newURL)
	q.Push(urlData{seed: newURL.Host, pUrl: newURL, dist: 0})
	inQueue[newURL.String()] = true

	for q.Size() > 0 {
		elem := q.Pop().(urlData)

		plainUrl := elem.pUrl.String()
		res, err := c.client.Get(plainUrl)
		if skip := LogResponse(plainUrl, res, err); skip {
			continue
		}

		reqUrl := ClearUrl(res.Request.URL)
		if _, ok := crawled[reqUrl]; ok {
			continue
		}

		discoverd++
		if discoverd > c.maxURL {
			return
		}

		// save new URL whose request went through
		crawled[reqUrl] = true
		elem.pUrl = res.Request.URL
		chURL <- elem
		if err = handler(res); err != nil {
			return
		}
	
		body := res.Body
		defer body.Close()
		doc, err := goquery.NewDocumentFromReader(body)
		if err != nil {
			fmt.Println("ERROR: can't read body")
			continue
		}
		
		selector := doc.Find("a")
		for i := range selector.Nodes {

			href, _ := selector.Eq(i).Attr("href")
			nextURL, err := newURL.Parse(href)
			if err != nil || (c.Restrict && nextURL.Host != newURL.Host) {
				continue
			}
		
			cleanURL := ClearUrl(nextURL)
			if _, ok := inQueue[cleanURL]; !ok {
				d := elem.dist
				if nextURL.Host != elem.pUrl.Host {
					d++
				}
				
				if d > c.Distance {
					continue
				}
							
				if q.Size() < c.maxQueued {
					data := urlData{seed: newURL.Host, pUrl: nextURL, dist: d}
					q.Push(data)
					// avoid duplicate URL in queue
					inQueue[cleanURL] = true
				}
			}
		}
	}
}

