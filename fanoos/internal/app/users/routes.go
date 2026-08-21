package users

import (
	"net/http"

	"github.com/amir1848/fanoos/internal/app/auth"
	"github.com/amir1848/fanoos/utils/convertutils"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/users")

	g.POST("/create", auth.RequireResource(ResourceCreate), createUser)
	g.POST("/update", auth.RequireResource(ResourceUpdate), updateUser)
	g.POST("/delete", auth.RequireResource(ResourceDelete), deleteUsers)
}

func createUser(ctx *gin.Context) {
	c := ctx.Request.Context()

	var user = UserSaveDTO{}

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := CreateUser(c, &User{
		ID:          user.ID,
		UserName:    user.UserName,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		IsAdmin:     user.IsAdmin,
		Password:    user.Password,
		FullName:    user.FirstName + " " + user.LastName,
		Email:       convertutils.ToNullString(user.Email),
		PhoneNumber: convertutils.ToNullString(user.PhoneNumber),
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": userID})
}

func updateUser(ctx *gin.Context) {
	c := ctx.Request.Context()

	var user = UserSaveDTO{}
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := UpdateUser(c, &User{
		ID:          user.ID,
		UserName:    user.UserName,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		IsAdmin:     user.IsAdmin,
		Password:    user.Password,
		FullName:    user.FirstName + " " + user.LastName,
		Email:       convertutils.ToNullString(user.Email),
		PhoneNumber: convertutils.ToNullString(user.PhoneNumber),
	}); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": user.ID})
}

func deleteUsers(ctx *gin.Context) {
	c := ctx.Request.Context()

	deleteIds := struct {
		Ids []int64
	}{}

	if err := ctx.ShouldBindJSON(&deleteIds); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := DeleteUser(c, deleteIds.Ids)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}
