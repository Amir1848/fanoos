package main

import (
	"github.com/amir1848/marketino/internal/app"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	api := r.Group("/api")
	{
		app.RegisterRoutes(api)
	}

	r.Static("/static", "./assets")
	r.NoRoute(func(c *gin.Context) {
		c.File("./assets/index.html")
	})

	r.Run(":8080")
}
