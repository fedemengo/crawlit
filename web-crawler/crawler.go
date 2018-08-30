package crawler

import (
	"fmt"
	"github.com/fedemengo/go-utility/data-structures/queue"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Handler callback type
type Handler func(res *http.Response) error

// Crawler represent an object to extrapolate link from website
type Crawler struct {
	URLs     []*url.URL
	Restrict bool
	Distance int
	timeout int
	maxURL	 int
	maxQueued int
	resultCh chan map[int][]string
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
			case id := <- quit:
				seeds--
				fmt.Println("COMPLETE", c.URLs[id].String())
			}
		}
		close(quit)
		c.resultCh <- result
		close(c.resultCh)
	}()
}

func (c Crawler) crawl(seedURL *url.URL, chURL chan *url.URL, handler Handler) {
	// one completed close channel
	defer func() {
		close(chURL)
	}()
	
	// initialize a client for each routine
	client := http.Client{
		Timeout: time.Duration(time.Duration(c.timeout) * time.Second),
	}

	discoverd := 0
	// keep track of the queued url
	inQueue := make(map[string]bool)
	// keep track of the crawled url (res.StatusCode == 200)
	crawled := make(map[string]bool)
	// keep track of distance to other hosts
	hostDist := make(map[string]int)
	q := queue.NewQueue()

	// resolve references if present
	seedURL = seedURL.ResolveReference(seedURL)
	// set initial distanc
	hostDist[seedURL.Host] = 0;
	// push the seed in queue
	q.Push(seedURL)
	inQueue[seedURL.String()] = true

	for q.Size() > 0 {
		currURL := q.Pop().(*url.URL)

		plainUrl := currURL.String()
		res, err := client.Get(plainUrl)
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
		currURL = res.Request.URL
		chURL <- currURL
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
			nextURL, err := currURL.Parse(href)
			if err != nil || (c.Restrict && nextURL.Host != seedURL.Host) {
				continue
			}
		
			// if url is already in queue, skip
			cleanURL := ClearUrl(nextURL)
			if _, ok := inQueue[cleanURL]; ok {
				continue
			}

			// check distance
			dist := hostDist[currURL.Host]
			if _, ok := hostDist[nextURL.Host]; !ok {
				dist++
				if dist > c.Distance {
					continue
				} else {
					hostDist[nextURL.Host] = dist
				}
			}
								
			if q.Size() < c.maxQueued {
				q.Push(nextURL)
				// avoid duplicate URL in queue
				inQueue[cleanURL] = true
			}
		}
	}
}

