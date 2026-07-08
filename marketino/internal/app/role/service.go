package role

import (
	"context"

	"github.com/amir1848/marketino/utils/dbutils"
	"gorm.io/gorm"
)

func CreateRole(ctx context.Context, role *Role) (int64, error) {
	var roleID int64

	err := dbutils.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(role).Error
		if err != nil {
			return err
		}

		roleID = role.ID

		return nil
	})
	if err != nil {
		return 0, err
	}

	return roleID, nil
}

func DeleteRole(ctx context.Context, ids []int64) error {
	err := dbutils.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`DELETE FROM roles WHERE id IN (?)`, ids).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
