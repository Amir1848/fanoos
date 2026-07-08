package roleresource

import (
	"context"

	"github.com/amir1848/marketino/utils/dbutils"
	"gorm.io/gorm"
)

func AssignResources(ctx context.Context, roleID int64, resourceIDs []int64) error {
	err := dbutils.Transaction(func(tx *gorm.DB) error {
		for _, resourceID := range resourceIDs {
			err := tx.Exec(`INSERT INTO role_resources (role_id, resource_id) VALUES (?, ?) ON CONFLICT DO NOTHING`, roleID, resourceID).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func RemoveResources(ctx context.Context, roleID int64, resourceIDs []int64) error {
	err := dbutils.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`DELETE FROM role_resources WHERE role_id = ? AND resource_id IN (?)`, roleID, resourceIDs).Error
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func GetRoleResources(ctx context.Context, roleID int64) ([]Resource, error) {
	var resources []Resource

	err := dbutils.Transaction(func(tx *gorm.DB) error {
		err := tx.Raw(`SELECT r.id, r.resource_key FROM resources r INNER JOIN role_resources rr ON r.id = rr.resource_id WHERE rr.role_id = ?`, roleID).Scan(&resources).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resources, nil
}
