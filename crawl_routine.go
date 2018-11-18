package crawlit

import (
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fedemengo/go-data-structures/queue"
)

func crawlURL(config CrawlConfig, id int, collect *[]string, done chan int, handler Handler) {
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

		body, err := goquery.NewDocumentFromReader(res.Body)
		res.Body.Close()
		if err != nil {
			continue
		}

		response := CrawlitResponse{
			URL:  res.Request.URL.String(),
			Body: body,
		}

		// if handler return an error consider stopping crawler or skipping the URL
		if err = handler(response); err != nil {
			switch err {
			case SkipURL:
				continue
			default:
				return
			}
		}

		// save new URL whose request went through
		crawled[cleanURL] = true
		elem.url, _ = url.Parse(cleanURL)
		*collect = append(*collect, elem.url.String())

		discovered++
		if discovered == config.MaxURLs {
			return
		}

		selector := body.Find("a")
		for i := range selector.Nodes {

			href, ok := selector.Eq(i).Attr("href")
			if !ok {
				continue
			}

			nextURL, err := elem.url.Parse(href)
			if err != nil {
				continue
			}

			if !ValidURL(config, elem, startURL, nextURL) {
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
