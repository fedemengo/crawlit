package crawler

import (
	"fmt"
	"go-scraper/url"
	"net"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type data struct {
	id  int
	url string
}

// Crawler represent an object to extrapolate link from website
type Crawler struct {
	Urls     []string
	Restrict bool
	Distance int
	resultCh chan map[string][]string
	host     []string
	crawled  map[string]bool
	client   http.Client
}

// NewCrawler creates a new Crawler instance
func NewCrawler(urls []string, restrict bool, distance, timeout int) *Crawler {
	c := new(Crawler)
	c.resultCh = make(chan map[string][]string)
	c.Distance = distance
	c.Restrict = restrict
	c.Urls = urls
	c.host = make([]string, len(urls))
	for i, u := range urls {
		h, _, _ := url.SplitURL(u)
		c.host[i] = h
	}
	c.crawled = make(map[string]bool)
	c.client = http.Client{
		Timeout: time.Duration(time.Duration(timeout) * time.Second),
	}
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
	chURLs := make(chan data)
	chDone := make(chan bool)

	for id, u := range c.Urls {
		h, p, _ := url.SplitURL(u)
		go c.crawl(h, p, id, c.Distance, chURLs, chDone)
	}

	// listen for result and termination
	go func() {
		for routines := len(c.Urls); routines > 0; {
			select {
			case d := <-chURLs:
				host := c.Urls[d.id]
				result[host] = append(result[host], d.url)
			case <-chDone:
				routines--
			}
		}
		c.resultCh <- result
		//close(c.resultCh)
	}()
}

func (c Crawler) crawl(host, path string, id, dist int, chURL chan data, chDone chan bool) {
	defer func() {
		if chDone != nil {
			chDone <- true
		}
	}()

	res, err := c.client.Get(host + path)
	if err != nil {
		if netError, ok := err.(net.Error); ok && netError.Timeout() {
			fmt.Println("ERROR: timeout on", "\""+host+path+"\"")
		} else {
			fmt.Println("ERROR: can't crawl", "\""+host+path+"\"")
		}
		return
	}

	body := res.Body
	defer body.Close()

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return
	}

	selector := doc.Find("a")
	for i := range selector.Nodes {
		href, _ := selector.Eq(i).Attr("href")

		discoveredURLs := url.CreateURL(host, href)
		for _, u := range discoveredURLs {
			newHost, newPath, err := url.SplitURL(u)
			if err != nil || (c.Restrict && newHost != c.host[id]) {
				continue
			}

			if _, ok := c.crawled[newHost+newPath]; !ok {
				chURL <- data{id: id, url: newHost + newPath}
				c.crawled[newHost+newPath] = true
				newDist := dist
				if newHost != host {
					newDist--
				}
				if newDist == 0 {
					continue
				}
				c.crawl(newHost, newPath, id, newDist, chURL, nil)
			}
		}
	}
}
