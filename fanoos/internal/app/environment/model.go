package environment

import (
	"database/sql"
	"time"
)

type Environment struct {
	ID          int64
	Name        string
	Description sql.NullString
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type EnvironmentSaveDTO struct {
	ID          int64
	Name        string
	Description string
}
