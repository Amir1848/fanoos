package app

import (
	assistanttools "github.com/amir1848/fanoos/internal/app/assistant/tools"
	"github.com/amir1848/fanoos/internal/app/auth"
	"github.com/amir1848/fanoos/internal/app/device"
	"github.com/amir1848/fanoos/internal/app/devicelink"
	"github.com/amir1848/fanoos/internal/app/environment"
	"github.com/amir1848/fanoos/internal/app/firmware"
	"github.com/amir1848/fanoos/internal/app/patoken"
	"github.com/amir1848/fanoos/internal/app/role"
	"github.com/amir1848/fanoos/internal/app/roleresource"
	"github.com/amir1848/fanoos/internal/app/users"
	"github.com/amir1848/fanoos/internal/app/userrole"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	auth.RegisterRoutes(r)

	protected := r.Group("")
	protected.Use(auth.AuthMiddleware())

	patoken.RegisterRoutes(protected)
	auth.RegisterProtectedRoutes(protected)
	role.RegisterRoutes(protected)
	roleresource.RegisterRoutes(protected)
	users.RegisterRoutes(protected)
	userrole.RegisterRoutes(protected)
	environment.RegisterRoutes(protected)
	device.RegisterRoutes(protected)
	devicelink.RegisterRoutes(protected)
	firmware.RegisterRoutes(protected)
	assistanttools.RegisterRoutes(protected)
}

func RegisterDataviews(r *gin.RouterGroup) {
	protected := r.Group("")
	protected.Use(auth.AuthMiddleware())

	role.RegisterDataviews(protected)
	roleresource.RegisterDataviews(protected)
	users.RegisterDataviews(protected)
	environment.RegisterDataviews(protected)
	device.RegisterDataviews(protected)
	devicelink.RegisterDataviews(protected)
}
