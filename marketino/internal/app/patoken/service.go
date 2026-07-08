package patoken

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/amir1848/marketino/utils/dbutils"
	"gorm.io/gorm"
)

const tokenPrefix = "mkt_pat_"

func generateToken() (rawToken string, tokenHash string, err error) {
	b := make([]byte, 64)
	_, err = rand.Read(b)
	if err != nil {
		return "", "", err
	}

	rawToken = tokenPrefix + hex.EncodeToString(b)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash = hex.EncodeToString(hash[:])

	return rawToken, tokenHash, nil
}

func CreateToken(ctx context.Context, userID int64, name string) (rawToken string, err error) {
	rawToken, tokenHash, err := generateToken()
	if err != nil {
		return "", err
	}

	err = dbutils.Transaction(func(tx *gorm.DB) error {
		token := PersonalAccessToken{
			UserID:    userID,
			Name:      name,
			TokenHash: tokenHash,
			CreatedAt: time.Now(),
		}

		return tx.Create(&token).Error
	})
	if err != nil {
		return "", err
	}

	return rawToken, nil
}

func RevokeTokens(ctx context.Context, userID int64, ids []int64) error {
	return dbutils.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(`DELETE FROM personal_access_tokens WHERE id IN (?) AND user_id = ?`, ids, userID).Error
	})
}

func ListTokens(ctx context.Context, userID int64) ([]TokenListItem, error) {
	var tokens []PersonalAccessToken
	var result []TokenListItem

	err := dbutils.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("user_id = ?", userID).Order("created_at DESC").Find(&tokens).Error
		if err != nil {
			return err
		}

		for _, t := range tokens {
			item := TokenListItem{
				ID:        t.ID,
				Name:      t.Name,
				CreatedAt: t.CreatedAt,
			}
			if t.LastUsedAt.Valid {
				item.LastUsedAt = &t.LastUsedAt.Time
			}
			result = append(result, item)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ValidateToken(ctx context.Context, rawToken string) (int64, error) {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	db := dbutils.GetDB()

	var token PersonalAccessToken
	err := db.Where("token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("invalid token")
		}
		return 0, err
	}

	db.Model(&token).Update("last_used_at", time.Now())

	return token.UserID, nil
}
