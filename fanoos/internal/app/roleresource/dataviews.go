package roleresource

import (
	"context"

	"github.com/amir1848/fanoos/internal/app/dataviews"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterDataviews(r *gin.RouterGroup) {
	dataviews.RegisterRoute(r, &dataviews.DataviewModel[struct {
		Id          int64  `json:"id"`
		ResourceKey string `json:"resourceKey"`
	}, struct{}]{
		DataviewKey: "all-resources",
		ResourceKey: ResourceView,
		Query: func(ctx context.Context, db *gorm.DB, _ *struct{}) (*gorm.DB, error) {
			return db.Table("resources").Select("*"), nil
		},
	})
}
