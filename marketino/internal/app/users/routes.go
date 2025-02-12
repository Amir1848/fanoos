package users

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/users")

	g.GET("", func(ctx *gin.Context) {

	})
}
