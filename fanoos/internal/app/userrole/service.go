package userrole

import (
	"context"

	"github.com/amir1848/fanoos/utils/dbutils"
	"gorm.io/gorm"
)

func AssignRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	err := dbutils.Transaction(func(tx *gorm.DB) error {
		for _, roleID := range roleIDs {
			err := tx.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES (?, ?) ON CONFLICT DO NOTHING`, userID, roleID).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func RemoveRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	err := dbutils.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`DELETE FROM user_roles WHERE user_id = ? AND role_id IN (?)`, userID, roleIDs).Error
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func ListRoles(ctx context.Context) ([]Role, error) {
	var roles []Role

	err := dbutils.Transaction(func(tx *gorm.DB) error {
		err := tx.Raw(`SELECT id, title, code FROM roles ORDER BY id`).Scan(&roles).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func GetUserRoles(ctx context.Context, userID int64) ([]Role, error) {
	var roles []Role

	err := dbutils.Transaction(func(tx *gorm.DB) error {
		err := tx.Raw(`SELECT r.id, r.title, r.code FROM roles r INNER JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id = ?`, userID).Scan(&roles).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return roles, nil
}
