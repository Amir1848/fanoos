package devicelink

import (
	"context"

	"github.com/amir1848/fanoos/internal/app/dataviews"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterDataviews(r *gin.RouterGroup) {
	dataviews.RegisterRoute(r, &dataviews.DataviewModel[struct {
		Id             int64  `json:"id"`
		SourceDeviceId int64  `json:"sourceDeviceId"`
		TargetDeviceId int64  `json:"targetDeviceId"`
		LinkType       int    `json:"linkType"`
	}, struct {
		EnvironmentID int64
	}]{
		DataviewKey: "all-device-links",
		Query: func(ctx context.Context, db *gorm.DB, params *struct {
			EnvironmentID int64
		}) (*gorm.DB, error) {
			q := db.Table("device_links dl").
				Select("dl.id, dl.source_device_id, dl.target_device_id, dl.link_type")
			if params != nil && params.EnvironmentID != 0 {
				q = q.Joins("JOIN devices d ON d.id = dl.source_device_id").
					Where("d.environment_id = ?", params.EnvironmentID)
			}
			return q, nil
		},
	})
}
