package devicelink

const (
	LinkTypeToggle = 1
)

var validLinkTypes = map[int]bool{
	LinkTypeToggle: true,
}

func IsValidLinkType(linkType int) bool {
	return validLinkTypes[linkType]
}
