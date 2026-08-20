package role

import (
	"net/http"

	"github.com/amir1848/fanoos/internal/app/auth"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/roles")

	g.POST("/create", auth.RequireResource(ResourceCreate), createRole)
	g.POST("/update", auth.RequireResource(ResourceUpdate), updateRole)
	g.POST("/delete", auth.RequireResource(ResourceDelete), deleteRoles)
}

func createRole(ctx *gin.Context) {
	c := ctx.Request.Context()

	var role = RoleSaveDTO{}

	if err := ctx.ShouldBindJSON(&role); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roleID, err := CreateRole(c, &Role{
		ID:    role.ID,
		Title: role.Title,
		Code:  role.Code,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": roleID})
}

func updateRole(ctx *gin.Context) {
	c := ctx.Request.Context()

	var role = RoleSaveDTO{}
	if err := ctx.ShouldBindJSON(&role); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := UpdateRole(c, &Role{
		ID:    role.ID,
		Title: role.Title,
		Code:  role.Code,
	}); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": role.ID})
}

func deleteRoles(ctx *gin.Context) {
	c := ctx.Request.Context()

	deleteIds := struct {
		Ids []int64
	}{}

	if err := ctx.ShouldBindJSON(&deleteIds); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deleted, err := DeleteRole(c, deleteIds.Ids)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": deleted})
}
