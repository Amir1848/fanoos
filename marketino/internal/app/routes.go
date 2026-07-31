package app

import (
	"github.com/amir1848/marketino/internal/app/auth"
	"github.com/amir1848/marketino/internal/app/device"
	"github.com/amir1848/marketino/internal/app/devicelink"
	"github.com/amir1848/marketino/internal/app/environment"
	"github.com/amir1848/marketino/internal/app/patoken"
	"github.com/amir1848/marketino/internal/app/role"
	"github.com/amir1848/marketino/internal/app/roleresource"
	"github.com/amir1848/marketino/internal/app/users"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	auth.RegisterRoutes(r)

	protected := r.Group("")
	protected.Use(auth.JWTMiddleware())

	patoken.RegisterRoutes(protected)
	role.RegisterRoutes(protected)
	roleresource.RegisterRoutes(protected)
	users.RegisterRoutes(protected)
	environment.RegisterRoutes(protected)
	device.RegisterRoutes(protected)
	devicelink.RegisterRoutes(protected)
}

func RegisterDataviews(r *gin.RouterGroup) {
	protected := r.Group("")
	protected.Use(auth.JWTMiddleware())

	role.RegisterDataviews(protected)
	roleresource.RegisterDataviews(protected)
	users.RegisterDataviews(protected)
	environment.RegisterDataviews(protected)
	device.RegisterDataviews(protected)
	devicelink.RegisterDataviews(protected)
}
