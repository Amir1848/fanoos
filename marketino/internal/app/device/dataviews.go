package device

import (
	"context"

	"github.com/amir1848/marketino/internal/app/dataviews"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterDataviews(r *gin.RouterGroup) {
	dataviews.RegisterRoute(r, &dataviews.DataviewModel[struct {
		Id            int64   `json:"id"`
		EnvironmentId int64   `json:"environmentId"`
		DeviceType    int     `json:"deviceType"`
		Name          string  `json:"name"`
		PositionX     float64 `json:"positionX"`
		PositionY     float64 `json:"positionY"`
		Rotation      float64 `json:"rotation"`
	}, struct {
		EnvironmentID int64
	}]{
		DataviewKey: "all-devices",
		Query: func(ctx context.Context, db *gorm.DB, params *struct {
			EnvironmentID int64
		}) (*gorm.DB, error) {
			q := db.Table("devices").Select("id, environment_id, device_type, name, position_x, position_y, rotation")
			if params != nil && params.EnvironmentID != 0 {
				q = q.Where("environment_id = ?", params.EnvironmentID)
			}
			return q, nil
		},
	})
}
