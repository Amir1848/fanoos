package role

import (
	"context"

	"github.com/amir1848/fanoos/internal/app/dataviews"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterDataviews(r *gin.RouterGroup) {
	dataviews.RegisterRoute(r, &dataviews.DataviewModel[struct {
		Id    int64  `json:"id"`
		Title string `json:"title"`
		Code  string `json:"code"`
	}, struct{}]{
		DataviewKey: "all-roles",
		Query: func(ctx context.Context, db *gorm.DB, _ *struct{}) (*gorm.DB, error) {
			return db.Table("roles").Select("*"), nil
		},
	})
}
