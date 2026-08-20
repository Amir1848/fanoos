package userrole

type Role struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Code  string `json:"code"`
}

type AssignDTO struct {
	UserID  int64
	RoleIDs []int64
}

type RemoveDTO struct {
	UserID  int64
	RoleIDs []int64
}
