package device

type Device struct {
	ID            int64
	EnvironmentID int64 `gorm:"column:environment_id"`
	DeviceType    int   `gorm:"column:device_type"`
	Name          string
	PositionX     float64 `gorm:"column:position_x"`
	PositionY     float64 `gorm:"column:position_y"`
	Rotation      float64
}

type DeviceSaveDTO struct {
	ID            int64
	EnvironmentID int64
	DeviceType    int
	Name          string
	PositionX     float64
	PositionY     float64
	Rotation      float64
}
