package patoken

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/personal-access-tokens")

	g.POST("/create", createToken)
	g.POST("/delete", deleteTokens)
	g.POST("/list", listTokens)
}

func createToken(ctx *gin.Context) {
	c := ctx.Request.Context()

	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var dto = CreateTokenDTO{}
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rawToken, err := CreateToken(c, userID.(int64), dto.Name)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": rawToken})
}

func deleteTokens(ctx *gin.Context) {
	c := ctx.Request.Context()

	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var dto = DeleteTokenDTO{}
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := RevokeTokens(c, userID.(int64), dto.IDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

func listTokens(ctx *gin.Context) {
	c := ctx.Request.Context()

	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	tokens, err := ListTokens(c, userID.(int64))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}
