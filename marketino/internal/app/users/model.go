package users

import (
	"database/sql"
	"time"
)

type User struct {
	ID          int64
	UserName    string
	FirstName   string
	LastName    string
	FullName    string
	Email       sql.NullString
	PhoneNumber sql.NullString
	UpdatedAt   time.Time
}

type UserSaveDTO struct {
	ID          int64
	UserName    string
	FirstName   string
	LastName    string
	Email       string
	PhoneNumber string
	UpdatedAt   time.Time
}
