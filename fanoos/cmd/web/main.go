package main

import (
	"mime"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"

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

	if os.Getenv("DEV_PROXY") != "" {
		proxyTarget := "http://localhost:4201"
		if target := os.Getenv("DEV_PROXY_TARGET"); target != "" {
			proxyTarget = target
		}
		targetURL, _ := url.Parse(proxyTarget)

		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		r.NoRoute(func(c *gin.Context) {
			c.Request.URL.Scheme = "http"
			c.Request.URL.Host = targetURL.Host
			proxy.ServeHTTP(c.Writer, c.Request)
		})

	} else {
		r.NoRoute(func(c *gin.Context) {
			name := filepath.Join("assets", filepath.Clean("/"+c.Request.URL.Path))
			if info, err := os.Stat(name); err == nil && !info.IsDir() {
				c.File(name)
				return
			}
			c.File(filepath.Join("assets", "index.html"))
		})
	}

	r.Run(":3000")
}
