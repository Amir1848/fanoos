package devicelink

import (
	"net/http"

	"github.com/amir1848/fanoos/internal/app/auth"
	"github.com/amir1848/fanoos/internal/app/environment"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/device-links")
	g.Use(auth.RequireResource(environment.ResourceManage))

	g.POST("/create", createDeviceLink)
	g.POST("/delete", deleteDeviceLinks)
}

func createDeviceLink(ctx *gin.Context) {
	c := ctx.Request.Context()

	var dto = DeviceLinkSaveDTO{}
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	linkID, err := CreateDeviceLink(c, &dto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": linkID})
}

func deleteDeviceLinks(ctx *gin.Context) {
	c := ctx.Request.Context()

	deleteIds := struct {
		Ids []int64
	}{}

	if err := ctx.ShouldBindJSON(&deleteIds); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DeleteDeviceLinks(c, deleteIds.Ids); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}
