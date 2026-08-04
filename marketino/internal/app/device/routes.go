package device

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/devices")

	g.POST("/create", createDevice)
	g.POST("/update", updateDevice)
	g.POST("/delete", deleteDevices)
}

func createDevice(ctx *gin.Context) {
	c := ctx.Request.Context()

	var dto = DeviceSaveDTO{}
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d, err := CreateDevice(c, &dto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": d.ID, "guid": d.GUID})
}

func updateDevice(ctx *gin.Context) {
	c := ctx.Request.Context()

	var dto = DeviceSaveDTO{}
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := UpdateDevice(c, &dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

func deleteDevices(ctx *gin.Context) {
	c := ctx.Request.Context()

	deleteIds := struct {
		Ids []int64
	}{}

	if err := ctx.ShouldBindJSON(&deleteIds); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DeleteDevices(c, deleteIds.Ids); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}
