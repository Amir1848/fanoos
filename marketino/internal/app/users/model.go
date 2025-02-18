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
	Password    string `gorm:"column:password_c"` 
	IsAdmin     bool
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
	Password    string 
	IsAdmin     bool
	UpdatedAt   time.Time
}
