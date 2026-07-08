package roleresource

type Resource struct {
	ID          int64
	ResourceKey string
}

type AssignDTO struct {
	RoleID      int64
	ResourceIDs []int64
}

type RemoveDTO struct {
	RoleID      int64
	ResourceIDs []int64
}
