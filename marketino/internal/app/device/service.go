package device

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amir1848/marketino/utils/dbutils"
	"github.com/amir1848/marketino/utils/mqttutils"
	"github.com/google/uuid"
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

type LightCommandDTO struct {
	DeviceID int64
	On       bool
}

// SendLightCommand looks up the device's environment and GUID, then
// publishes an `{"on": bool}` command to its MQTT cmd topic (see
// docs/architucture.md §8). The device must be a light.
func SendLightCommand(ctx context.Context, dto *LightCommandDTO) error {
	var target struct {
		EnvironmentID int64
		GUID          uuid.UUID
		DeviceType    int
	}

	err := dbutils.GetDB().WithContext(ctx).
		Table("devices").
		Select("guid, environment_id, device_type").
		Where("id = ?", dto.DeviceID).
		Take(&target).Error
	if err != nil {
		return fmt.Errorf("device %d not found: %w", dto.DeviceID, err)
	}

	if target.DeviceType != TypeLight {
		return fmt.Errorf("device %d is not a light", dto.DeviceID)
	}

	payload, err := json.Marshal(struct {
		On bool `json:"on"`
	}{On: dto.On})
	if err != nil {
		return err
	}

	topic := mqttutils.LightCmdTopic(target.EnvironmentID, target.GUID)
	if err := mqttutils.Publish(topic, payload); err != nil {
		return fmt.Errorf("publish to %s failed: %w", topic, err)
	}

	return nil
}
