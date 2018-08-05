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

type data struct {
	id  int
	url string
}

type urlData struct {
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
	host     []string
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
	c.host = make([]string, len(urls))
	for i, u := range urls {
		h, _, _ := url.SplitURL(u)
		c.host[i] = h
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
		go c.crawl(h, p, id, chURLs, chDone)
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

func (c Crawler) crawl(baseHost, basePath string, id int, chURL chan data, chDone chan bool) {
	defer func() {
		chDone <- true
	}()

	discoverd := 0
	crawled := make(map[string]bool)
	q := queue.NewQueue()
	q.Push(urlData{host: baseHost, path: basePath, dist: 0})

	for q.Size() > 0 {
		elem := q.Pop().(urlData)
		host, path, dist := elem.host, elem.path, elem.dist

		res, err := c.client.Get(host + path)
		if err != nil {
			if netError, ok := err.(net.Error); ok && netError.Timeout() {
				fmt.Println("ERROR: timeout on", "\""+host+path+"\"")
			} else {
				fmt.Println("ERROR: can't crawl", "\""+host+path+"\"")
			}
			continue
		} else if res.StatusCode == 404 {
			fmt.Println("Code 404: skipping", "\""+host+path+"\"...")
			continue
		}

		body := res.Body
		defer body.Close()
	
		doc, err := goquery.NewDocumentFromReader(body)
		if err != nil {
			continue
		}
		
		selector := doc.Find("a")
		for i := range selector.Nodes {
			href, _ := selector.Eq(i).Attr("href")

			discoverdURLs := url.CreateURL(host, href)
			for _, u := range discoverdURLs {

				newHost, newPath, err := url.SplitURL(u)
				if err != nil || (c.Restrict && newHost != c.host[id]) {
					continue
				}
	
				if _, present := crawled[newHost+newPath]; !present {
					if newHost != host {
						dist++
					}
	
					if dist > c.Distance {
						continue
					}
	
					discoverd++
					if discoverd > c.maxURL {
						return
					}

					crawled[newHost+newPath] = true
					chURL <- data{id: id, url: newHost + newPath}
					q.Push(urlData{host: newHost, path: newPath, dist: dist})

				}
			}
		}
	}
}
