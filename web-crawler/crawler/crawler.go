package crawler

import (
	"fmt"
	"github.com/fedemengo/search-engine/web-crawler/url"
	"github.com/fedemengo/go-utility/data-structures/queue"
	"net"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type urlData struct {
	seedHost string
	host string
	path string
	dist int
}

// Crawler represent an object to extrapolate link from website
type Crawler struct {
	Urls     []string
	Restrict bool
	Distance int
	maxURL	 int
	client   http.Client
	resultCh chan map[string][]string
}

// NewCrawler creates a new Crawler instance
func NewCrawler(urls []string, restrict bool, distance, timeout, maxURL int) *Crawler {
	c := new(Crawler)
	c.Urls = urls
	c.Restrict = restrict
	c.Distance = distance
	c.maxURL = maxURL
	c.client = http.Client{
		Timeout: time.Duration(time.Duration(timeout) * time.Second),
	}
	c.resultCh = make(chan map[string][]string)
	return c
}

// Result will return the result of the crawling
func (c *Crawler) Result() (urls map[string][]string) {
	urls = <-c.resultCh
	return
}

// Crawl is the public method used to start the crawling
func (c *Crawler) Crawl() {
	result := make(map[string][]string)
	chURLs := make(chan urlData)

	for _, u := range c.Urls {
		h, p, _ := url.SplitURL(u)
		go c.crawl(h, p, chURLs)
	}

	// listen for result and termination
	go func() {
		for routines := len(c.Urls); routines > 0; {
			select {
			case data, ok := <-chURLs:
				if ok {
					result[data.seedHost] = append(result[data.seedHost], data.host + data.path)
				} else {
					routines--
				} 
			}
		}
		c.resultCh <- result
		//close(c.resultCh)
	}()
}

func (c Crawler) crawl(baseHost, basePath string, chURL chan urlData) {
	defer func() {
		close(chURL)
	}()

	discoverd := 0
	inQueue := make(map[string]bool)
	q := queue.NewQueue()
	q.Push(urlData{seedHost: baseHost, host: baseHost, path: basePath, dist: 0})
	inQueue[baseHost + basePath] = true

	for q.Size() > 0 {
		elem := q.Pop().(urlData)

		res, err := c.client.Get(elem.host + elem.path)
		if err != nil {
			if netError, ok := err.(net.Error); ok && netError.Timeout() {
				fmt.Println("ERROR: TIMEOUT on", "\""+elem.host+elem.path+"\"")
			} else {
				fmt.Println("ERROR: can't crawl", "\""+elem.host+elem.path+"\"")
			}
			continue
		} else if res.StatusCode == 404 {
			fmt.Println("Code 404: skipping", "\""+elem.host+elem.path+"\"")
			continue
		}

		// save new URL
		chURL <- elem

		body := res.Body
		defer body.Close()
	
		doc, err := goquery.NewDocumentFromReader(body)
		if err != nil {
			continue
		}
		
		selector := doc.Find("a")
		for i := range selector.Nodes {

			href, _ := selector.Eq(i).Attr("href")
			discoverdURLs := url.CreateURL(elem.host, href)
			for _, u := range discoverdURLs {
				
				newHost, newPath, err := url.SplitURL(u)
				if err != nil || (c.Restrict && newHost != baseHost) {
					continue
				}
				
				if _, ok := inQueue[newHost+newPath]; !ok {
					dist := elem.dist
					if newHost != elem.host {
						dist++
					}
					
					if dist > c.Distance {
						continue
					}
					
					discoverd++
					if discoverd > c.maxURL {
						return
					}
					data := urlData{seedHost: baseHost, host: newHost, path: newPath, dist: dist}
					q.Push(data)
					inQueue[newHost + newPath] = true
				}
			}
		}
	}
}
