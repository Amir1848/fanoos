package auth

import (
	"context"

	"github.com/amir1848/fanoos/utils/dbutils"
)

// UserResourceKeys returns every resource key the user has access to:
// admins implicitly hold all resources, everyone else only the ones granted
// through their roles.
func UserResourceKeys(ctx context.Context, userID int64) ([]string, error) {
	db := dbutils.GetDB()

	var isAdmin bool
	err := db.Raw(`SELECT is_admin FROM users WHERE id = ?`, userID).Scan(&isAdmin).Error
	if err != nil {
		return nil, err
	}

	if isAdmin {
		var keys []string
		err := db.WithContext(ctx).Raw(`SELECT resource_key FROM resources ORDER BY id`).Scan(&keys).Error
		return keys, err
	}

	var keys []string
	err = db.WithContext(ctx).Raw(`
		SELECT DISTINCT r.resource_key
		FROM user_roles ur
		JOIN role_resources rr ON ur.role_id = rr.role_id
		JOIN resources r ON rr.resource_id = r.id
		WHERE ur.user_id = ?
		ORDER BY r.resource_key
	`, userID).Scan(&keys).Error
	if err != nil {
		return nil, err
	}

	return keys, nil
}
