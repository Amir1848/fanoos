package role

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/roles")

	g.POST("/create", createRole)
	g.POST("/delete", deleteRoles)
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

func deleteRoles(ctx *gin.Context) {
	c := ctx.Request.Context()

	deleteIds := struct {
		Ids []int64
	}{}

	if err := ctx.ShouldBindJSON(&deleteIds); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := DeleteRole(c, deleteIds.Ids)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}
