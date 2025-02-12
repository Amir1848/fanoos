package app

import (
	"github.com/amir1848/marketino/internal/app/users"
	"github.com/gin-gonic/gin"
)


func RegisterRoutes(r *gin.RouterGroup){
	users.RegisterRoutes(r)
}