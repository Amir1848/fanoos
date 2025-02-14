package main

import (
	"net/http/httputil"
	"net/url"

	"github.com/amir1848/marketino/internal/app"
	_ "github.com/amir1848/marketino/utils/dbutils"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	api := r.Group("/api")
	{
		app.RegisterRoutes(api)
	}

	if true { //check with env
		proxyTarget := "http://localhost:4201"
		targetURL, _ := url.Parse(proxyTarget)

		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		r.NoRoute(func(c *gin.Context) {
			c.Request.URL.Scheme = "http"
			c.Request.URL.Host = targetURL.Host
			proxy.ServeHTTP(c.Writer, c.Request)
		})

	} else {
		r.Static("/static", "./assets")
		r.NoRoute(func(c *gin.Context) {
			c.File("./assets/index.html")
		})
	}

	r.Run(":3000")
}
