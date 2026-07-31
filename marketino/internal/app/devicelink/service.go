package devicelink

import (
	"context"
	"fmt"

	"github.com/amir1848/marketino/internal/app/device"
	"github.com/amir1848/marketino/utils/dbutils"
	"gorm.io/gorm"
)

func CreateDeviceLink(ctx context.Context, dto *DeviceLinkSaveDTO) (int64, error) {
	var linkID int64

	err := dbutils.Transaction(func(tx *gorm.DB) error {
		var sourceType, targetType int
		if err := tx.Table("devices").Select("device_type").Where("id = ?", dto.SourceDeviceID).Scan(&sourceType).Error; err != nil {
			return err
		}
		if err := tx.Table("devices").Select("device_type").Where("id = ?", dto.TargetDeviceID).Scan(&targetType).Error; err != nil {
			return err
		}

		if !isLightControllerPair(sourceType, targetType) {
			return fmt.Errorf("a link must connect a light and a controller")
		}

		linkType := dto.LinkType
		if linkType == "" {
			linkType = "toggle"
		}

		link := DeviceLink{
			SourceDeviceID: dto.SourceDeviceID,
			TargetDeviceID: dto.TargetDeviceID,
			LinkType:       linkType,
		}

		if err := tx.Create(&link).Error; err != nil {
			return err
		}

		linkID = link.ID

		return nil
	})
	if err != nil {
		return 0, err
	}

	return linkID, nil
}

func DeleteDeviceLinks(ctx context.Context, ids []int64) error {
	return dbutils.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(`DELETE FROM device_links WHERE id IN (?)`, ids).Error
	})
}

func isLightControllerPair(a, b int) bool {
	return (a == device.TypeLight && b == device.TypeController) || (a == device.TypeController && b == device.TypeLight)
}
