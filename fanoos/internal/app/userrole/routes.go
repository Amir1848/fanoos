package userrole

import (
	"net/http"

	"github.com/amir1848/fanoos/internal/app/auth"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/user-roles")

	g.POST("/assign", auth.RequireResource(ResourceAssign), assignRoles)
	g.POST("/remove", auth.RequireResource(ResourceAssign), removeRoles)
	g.POST("/list", auth.RequireResource(ResourceView), getUserRoles)
	g.POST("/roles", auth.RequireResource(ResourceView), getAllRoles)
}

func assignRoles(ctx *gin.Context) {
	c := ctx.Request.Context()

	var dto = AssignDTO{}
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := AssignRoles(c, dto.UserID, dto.RoleIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

func removeRoles(ctx *gin.Context) {
	c := ctx.Request.Context()

	var dto = RemoveDTO{}
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := RemoveRoles(c, dto.UserID, dto.RoleIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

func getUserRoles(ctx *gin.Context) {
	c := ctx.Request.Context()

	var req struct {
		UserID int64
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roles, err := GetUserRoles(c, req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, roles)
}

func getAllRoles(ctx *gin.Context) {
	roles, err := ListRoles(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, roles)
}
