package app

import (
	"github.com/amir1848/marketino/internal/app/auth"
	"github.com/amir1848/marketino/internal/app/patoken"
	"github.com/amir1848/marketino/internal/app/role"
	"github.com/amir1848/marketino/internal/app/roleresource"
	"github.com/amir1848/marketino/internal/app/users"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	auth.RegisterRoutes(r)
	patoken.RegisterRoutes(r)
	role.RegisterRoutes(r)
	roleresource.RegisterRoutes(r)
	users.RegisterRoutes(r)
}

func RegisterDataviews(r *gin.RouterGroup) {
	role.RegisterDataviews(r)
	roleresource.RegisterDataviews(r)
	users.RegisterDataviews(r)
}
