package users

import (
	"context"

	"github.com/amir1848/marketino/internal/app/dataviews"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterDataviews(r *gin.RouterGroup) {
	dataviews.RegisterRoute(r, &dataviews.DataviewModel[struct {
		FirstName string
		LastName  string
	}, struct{}]{
		DataviewKey: "all-users",
		Query: func(ctx context.Context, db *gorm.DB, _ *struct{}) (*gorm.DB, error) {
			return db.Table("users").Select("*"), nil
		},
	})
}
