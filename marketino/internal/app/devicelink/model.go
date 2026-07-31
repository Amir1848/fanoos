package devicelink

type DeviceLink struct {
	ID             int64
	SourceDeviceID int64  `gorm:"column:source_device_id"`
	TargetDeviceID int64  `gorm:"column:target_device_id"`
	LinkType       string `gorm:"column:link_type"`
}

type DeviceLinkSaveDTO struct {
	SourceDeviceID int64
	TargetDeviceID int64
	LinkType       string
}
