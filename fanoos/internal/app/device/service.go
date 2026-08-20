package device

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/amir1848/fanoos/utils/dbutils"
	"github.com/amir1848/fanoos/utils/mqttutils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func init() {
	mqttutils.SetStateHandler(handleDeviceStateUpdate)
}

// handleDeviceStateUpdate persists a device's reported state to the
// device_states table so any backend instance can read it back (state
// can't live in memory once more than one instance is running - see
// docs/architucture.md §8). A nil field in update means the MQTT message
// didn't mention it (e.g. a last-will "offline" message only carries
// Online), so the previously stored value for that field is kept.
func handleDeviceStateUpdate(deviceGUID uuid.UUID, update mqttutils.StateUpdate) {
	db := dbutils.GetDB()

	var existing struct {
		IsOn   bool
		Online bool
	}
	err := db.Table("device_states").Select("is_on, online").Where("device_guid = ?", deviceGUID).Take(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("device: reading previous state for %s failed: %v", deviceGUID, err)
		return
	}

	on := existing.IsOn
	if update.On != nil {
		on = *update.On
	}
	online := existing.Online
	if update.Online != nil {
		online = *update.Online
	}

	err = db.Exec(`
		INSERT INTO device_states (device_guid, is_on, online, ts, updated_at)
		VALUES (?, ?, ?, ?, now())
		ON CONFLICT (device_guid) DO UPDATE SET
			is_on = EXCLUDED.is_on, online = EXCLUDED.online, ts = EXCLUDED.ts, updated_at = EXCLUDED.updated_at
	`, deviceGUID, on, online, update.Ts).Error
	if err != nil {
		log.Printf("device: storing state for %s failed: %v", deviceGUID, err)
	}
}

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

type LightStateDTO struct {
	On     bool  `json:"on"`
	Online bool  `json:"online"`
	Ts     int64 `json:"ts"`
	Known  bool  `json:"known"`
}

type ControllerStateDTO struct {
	Online bool  `json:"online"`
	Ts     int64 `json:"ts"`
	Known  bool  `json:"known"`
}

// GetLightState returns the last state reported by a light over MQTT,
// persisted to the device_states table by handleDeviceStateUpdate (see
// docs/architucture.md §8). Known is false if no device_states row exists
// yet for this device.
func GetLightState(ctx context.Context, deviceID int64) (*LightStateDTO, error) {
	var target struct {
		GUID       uuid.UUID
		DeviceType int
	}

	err := dbutils.GetDB().WithContext(ctx).
		Table("devices").
		Select("guid, device_type").
		Where("id = ?", deviceID).
		Take(&target).Error
	if err != nil {
		return nil, fmt.Errorf("device %d not found: %w", deviceID, err)
	}

	if target.DeviceType != TypeLight {
		return nil, fmt.Errorf("device %d is not a light", deviceID)
	}

	var row struct {
		IsOn   bool
		Online bool
		Ts     int64
	}
	err = dbutils.GetDB().WithContext(ctx).
		Table("device_states").
		Select("is_on, online, ts").
		Where("device_guid = ?", target.GUID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &LightStateDTO{Known: false}, nil
	}
	if err != nil {
		return nil, err
	}

	return &LightStateDTO{On: row.IsOn, Online: row.Online, Ts: row.Ts, Known: true}, nil
}

// GetControllerState returns the last online/offline state reported by a
// controller over MQTT, persisted to the device_states table by
// handleDeviceStateUpdate (see docs/architucture.md §8). Known is false
// if no device_states row exists yet for this device. Controllers have no
// "on" concept — `is_on` is meaningless for them and intentionally omitted.
func GetControllerState(ctx context.Context, deviceID int64) (*ControllerStateDTO, error) {
	var target struct {
		GUID       uuid.UUID
		DeviceType int
	}

	err := dbutils.GetDB().WithContext(ctx).
		Table("devices").
		Select("guid, device_type").
		Where("id = ?", deviceID).
		Take(&target).Error
	if err != nil {
		return nil, fmt.Errorf("device %d not found: %w", deviceID, err)
	}

	if target.DeviceType != TypeController {
		return nil, fmt.Errorf("device %d is not a controller", deviceID)
	}

	var row struct {
		Online bool
		Ts     int64
	}
	err = dbutils.GetDB().WithContext(ctx).
		Table("device_states").
		Select("online, ts").
		Where("device_guid = ?", target.GUID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &ControllerStateDTO{Known: false}, nil
	}
	if err != nil {
		return nil, err
	}

	return &ControllerStateDTO{Online: row.Online, Ts: row.Ts, Known: true}, nil
}
