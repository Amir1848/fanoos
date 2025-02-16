package app

import (
	"github.com/amir1848/marketino/internal/app/auth"
	"github.com/amir1848/marketino/internal/app/users"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	auth.RegisterRoutes(r)
	users.RegisterRoutes(r)
}

func RegisterDataviews(r *gin.RouterGroup) {
	users.RegisterDataviews(r)
}
