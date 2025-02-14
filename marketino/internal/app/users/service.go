package users

import (
	"context"

	"github.com/amir1848/marketino/utils/dbutils"
	"gorm.io/gorm"
)

func CreateUser(ctx context.Context, user *User) (int64, error) {
	var userID int64

	err := dbutils.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(user).Error
		if err != nil {
			return err
		}

		userID = user.ID

		return nil
	})
	if err != nil {
		return 0, err
	}

	return userID, nil
}
