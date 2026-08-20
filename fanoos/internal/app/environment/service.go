package environment

import (
	"context"
	"time"

	"github.com/amir1848/fanoos/utils/convertutils"
	"github.com/amir1848/fanoos/utils/dbutils"
	"gorm.io/gorm"
)

func CreateEnvironment(ctx context.Context, dto *EnvironmentSaveDTO) (int64, error) {
	var environmentID int64

	err := dbutils.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		env := Environment{
			Name:        dto.Name,
			Description: convertutils.ToNullString(dto.Description),
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := tx.Create(&env).Error; err != nil {
			return err
		}

		environmentID = env.ID

		return nil
	})
	if err != nil {
		return 0, err
	}

	return environmentID, nil
}

func UpdateEnvironment(ctx context.Context, dto *EnvironmentSaveDTO) error {
	return dbutils.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(
			`UPDATE environments SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
			dto.Name, convertutils.ToNullString(dto.Description), time.Now(), dto.ID,
		).Error
	})
}

func DeleteEnvironments(ctx context.Context, ids []int64) error {
	return dbutils.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(`DELETE FROM environments WHERE id IN (?)`, ids).Error
	})
}
