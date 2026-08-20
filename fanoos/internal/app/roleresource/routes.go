package roleresource

import (
	"net/http"

	"github.com/amir1848/fanoos/internal/app/auth"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/role-resources")

	g.POST("/assign", auth.RequireResource(ResourceAssign), assignResources)
	g.POST("/remove", auth.RequireResource(ResourceAssign), removeResources)
	g.POST("/list", auth.RequireResource(ResourceView), getRoleResources)
}

func assignResources(ctx *gin.Context) {
	c := ctx.Request.Context()

	var dto = AssignDTO{}
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := AssignResources(c, dto.RoleID, dto.ResourceIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

func removeResources(ctx *gin.Context) {
	c := ctx.Request.Context()

	var dto = RemoveDTO{}
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := RemoveResources(c, dto.RoleID, dto.ResourceIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

func getRoleResources(ctx *gin.Context) {
	c := ctx.Request.Context()

	var req struct {
		RoleID int64
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resources, err := GetRoleResources(c, req.RoleID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resources)
}
