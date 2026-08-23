package auth

import (
	"context"

	"github.com/amir1848/fanoos/utils/dbutils"
)

// UserResourceKeys returns the username and every resource key the user
// has access to: admins implicitly hold all resources, everyone else only
// the ones granted through their roles.
func UserResourceKeys(ctx context.Context, userID int64) (keys []string, username string, err error) {
	db := dbutils.GetDB()

	var row struct {
		UserName string
		IsAdmin  bool
	}
	err = db.WithContext(ctx).Raw(`SELECT user_name, is_admin FROM users WHERE id = ?`, userID).Scan(&row).Error
	if err != nil {
		return nil, "", err
	}
	if row.UserName == "" {
		return nil, "", errInvalidSession
	}

	if row.IsAdmin {
		err = db.WithContext(ctx).Raw(`SELECT resource_key FROM resources ORDER BY id`).Scan(&keys).Error
		return keys, row.UserName, err
	}

	err = db.WithContext(ctx).Raw(`
		SELECT DISTINCT r.resource_key
		FROM user_roles ur
		JOIN role_resources rr ON ur.role_id = rr.role_id
		JOIN resources r ON rr.resource_id = r.id
		WHERE ur.user_id = ?
		ORDER BY r.resource_key
	`, userID).Scan(&keys).Error
	if err != nil {
		return nil, "", err
	}

	return keys, row.UserName, nil
}
