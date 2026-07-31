package environment

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/environments")

	g.POST("/create", createEnvironment)
	g.POST("/update", updateEnvironment)
	g.POST("/delete", deleteEnvironments)
}

func createEnvironment(ctx *gin.Context) {
	c := ctx.Request.Context()

	var dto = EnvironmentSaveDTO{}
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	environmentID, err := CreateEnvironment(c, &dto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": environmentID})
}

func updateEnvironment(ctx *gin.Context) {
	c := ctx.Request.Context()

	var dto = EnvironmentSaveDTO{}
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := UpdateEnvironment(c, &dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

func deleteEnvironments(ctx *gin.Context) {
	c := ctx.Request.Context()

	deleteIds := struct {
		Ids []int64
	}{}

	if err := ctx.ShouldBindJSON(&deleteIds); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DeleteEnvironments(c, deleteIds.Ids); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}
