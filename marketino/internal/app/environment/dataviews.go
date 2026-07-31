package environment

import (
	"context"

	"github.com/amir1848/marketino/internal/app/dataviews"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterDataviews(r *gin.RouterGroup) {
	dataviews.RegisterRoute(r, &dataviews.DataviewModel[struct {
		Id          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}, struct{}]{
		DataviewKey: "all-environments",
		Query: func(ctx context.Context, db *gorm.DB, _ *struct{}) (*gorm.DB, error) {
			return db.Table("environments").Select("id, name, description"), nil
		},
	})
}
