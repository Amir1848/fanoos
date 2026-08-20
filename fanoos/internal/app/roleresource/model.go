package roleresource

type Resource struct {
	ID          int64  `json:"id"`
	ResourceKey string `json:"resourceKey"`
}

type AssignDTO struct {
	RoleID      int64
	ResourceIDs []int64
}

type RemoveDTO struct {
	RoleID      int64
	ResourceIDs []int64
}
