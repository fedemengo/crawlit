package main

import (
	"github.com/fedemengo/search-engine/backend/api"
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func init() {
	router = gin.Default()
	router.Delims("${", "}")
	router.LoadHTMLGlob("./public/*")
}

func main() {

	router.GET("/", api.SearchHandler)
	router.Run(":4000")
}
