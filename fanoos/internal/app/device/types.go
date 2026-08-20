package device

const (
	TypeLight      = 1
	TypeController = 2
)

var validTypes = map[int]bool{
	TypeLight:      true,
	TypeController: true,
}

func IsValidType(deviceType int) bool {
	return validTypes[deviceType]
}
