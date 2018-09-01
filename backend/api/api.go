package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type resData struct {
	Url  string
	Name string
}

func getData(query string) []resData {
	const maxURL = 30
	n, err := strconv.Atoi(query)
	fmt.Println(n)
	var urls []resData
	if err == nil {
		if n > maxURL {
			n = maxURL
		}
		urls = make([]resData, n)
		for i := 0; i < n; i++ {
			urls[i] = resData{
				Url:  "www.myurl-" + strconv.Itoa(i) + ".com",
				Name: "myurl-" + strconv.Itoa(i),
			}
		}
	}
	return urls
}

// SearchHandler handle the request of the Search endpoint
func SearchHandler(c *gin.Context) {
	query := c.Query("q")
	data := getData(query)

	c.HTML(http.StatusOK, "search.html", gin.H{
		"urls": data,
	})
}

// CrawlHandler handle the request of the Crawl endpoint
func CrawlHandler(res http.ResponseWriter, req *http.Request) {

}

// InfoHandler handle the request of the Info endpoint
func InfoHandler(res http.ResponseWriter, req *http.Request) {

}
