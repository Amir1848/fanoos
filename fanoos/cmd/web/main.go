package main

import (
	"mime"
	"net/http/httputil"
	"net/url"

	"github.com/amir1848/fanoos/internal/app"
	_ "github.com/amir1848/fanoos/utils/dbutils"
	"github.com/gin-gonic/gin"
)

func main() {
	// Ensure the PWA manifest gets a JSON MIME type regardless of the host's
	// mime.types (browsers reject manifests served as octet-stream).
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")

	r := gin.Default()

	api := r.Group("/api")
	{
		app.RegisterRoutes(api)
		app.RegisterDataviews(api)
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
