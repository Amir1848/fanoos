package device

import (
	"net/http"
	"strconv"

	"github.com/amir1848/fanoos/internal/app/auth"
	"github.com/amir1848/fanoos/internal/app/environment"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/devices")
	g.Use(auth.RequireResource(environment.ResourceManage))

	g.POST("/create", createDevice)
	g.POST("/update", updateDevice)
	g.POST("/delete", deleteDevices)
	g.POST("/command/light", sendLightCommand)
	g.GET("/:id/state/light", getLightState)
	g.GET("/:id/state/controller", getControllerState)
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

func sendLightCommand(ctx *gin.Context) {
	c := ctx.Request.Context()

	var dto LightCommandDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := SendLightCommand(c, &dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func getLightState(ctx *gin.Context) {
	c := ctx.Request.Context()

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	state, err := GetLightState(c, id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, state)
}

func getControllerState(ctx *gin.Context) {
	c := ctx.Request.Context()

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	state, err := GetControllerState(c, id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, state)
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
