package device

import (
	"context"
	"fmt"

	"github.com/amir1848/marketino/utils/dbutils"
	"gorm.io/gorm"
)

func CreateDevice(ctx context.Context, dto *DeviceSaveDTO) (*Device, error) {
	if !IsValidType(dto.DeviceType) {
		return nil, fmt.Errorf("invalid device type: %d", dto.DeviceType)
	}

	d := Device{
		EnvironmentID: dto.EnvironmentID,
		DeviceType:    dto.DeviceType,
		Name:          dto.Name,
		PositionX:     dto.PositionX,
		PositionY:     dto.PositionY,
		Rotation:      dto.Rotation,
	}

	err := dbutils.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&d).Error
	})
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func UpdateDevice(ctx context.Context, dto *DeviceSaveDTO) error {
	if !IsValidType(dto.DeviceType) {
		return fmt.Errorf("invalid device type: %d", dto.DeviceType)
	}

	return dbutils.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(
			`UPDATE devices SET device_type = ?, name = ?, position_x = ?, position_y = ?, rotation = ? WHERE id = ?`,
			dto.DeviceType, dto.Name, dto.PositionX, dto.PositionY, dto.Rotation, dto.ID,
		).Error
	})
}

func DeleteDevices(ctx context.Context, ids []int64) error {
	return dbutils.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(`DELETE FROM devices WHERE id IN (?)`, ids).Error
	})
}
