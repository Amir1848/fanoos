package role

import (
	"context"

	"github.com/amir1848/fanoos/utils/dbutils"
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

func UpdateRole(ctx context.Context, role *Role) error {
	return dbutils.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(
			`UPDATE roles SET title = ?, code = ? WHERE id = ?`,
			role.Title, role.Code, role.ID,
		).Error
	})
}

func DeleteRole(ctx context.Context, ids []int64) (int64, error) {
	var deleted int64

	err := dbutils.Transaction(func(tx *gorm.DB) error {
		result := tx.Exec(`DELETE FROM roles WHERE id IN (?)`, ids)
		if result.Error != nil {
			return result.Error
		}

		deleted = result.RowsAffected

		return nil
	})
	if err != nil {
		return 0, err
	}

	return deleted, nil
}
