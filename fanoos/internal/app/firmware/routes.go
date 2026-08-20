package firmware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/devices")

	g.GET("/:id/firmware", getFirmware)
}

// getFirmware serves the generated Arduino source + README for a controller
// device (the "get source code" action in the environment editor).
func getFirmware(ctx *gin.Context) {
	c := ctx.Request.Context()

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	fw, err := GenerateForController(c, id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, fw)
}
