package users

import (
	"context"

	"github.com/amir1848/fanoos/internal/app/dataviews"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterDataviews(r *gin.RouterGroup) {
	dataviews.RegisterRoute(r, &dataviews.DataviewModel[struct {
		Id          int64  `json:"id"`
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		UserName    string `json:"userName"`
		PhoneNumber string `json:"phoneNumber"`
		Email       string `json:"email"`
		IsAdmin     bool   `json:"isAdmin"`
	}, struct{}]{
		DataviewKey: "all-users",
		Query: func(ctx context.Context, db *gorm.DB, _ *struct{}) (*gorm.DB, error) {
			return db.Table("users").Select("*"), nil
		},
	})
}
