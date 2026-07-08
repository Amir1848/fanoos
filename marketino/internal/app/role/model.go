package role

type Role struct {
	ID    int64
	Title string
	Code  string
}

type RoleSaveDTO struct {
	ID    int64
	Title string
	Code  string
}
