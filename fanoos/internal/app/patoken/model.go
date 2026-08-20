package patoken

import (
	"database/sql"
	"time"
)

type PersonalAccessToken struct {
	ID         int64
	UserID     int64          `gorm:"column:user_id"`
	Name       string
	TokenHash  string         `gorm:"column:token_hash"`
	LastUsedAt sql.NullTime   `gorm:"column:last_used_at"`
	CreatedAt  time.Time      `gorm:"column:created_at"`
}

type CreateTokenDTO struct {
	Name string
}

type DeleteTokenDTO struct {
	IDs []int64
}

type TokenListItem struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}
