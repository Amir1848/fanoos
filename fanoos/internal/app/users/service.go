package users

import (
	"context"

	"github.com/amir1848/fanoos/internal/app/auth"
	"github.com/amir1848/fanoos/utils/dbutils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func init() {
	auth.RegisterGetUserData(getUser)
}

func CreateUser(ctx context.Context, user *User) (int64, error) {
	var userID int64

	err := dbutils.Transaction(func(tx *gorm.DB) error {
		err := validateUser(ctx, user)
		if err != nil {
			return err
		}

		err = prepareData(ctx, user)
		if err != nil {
			return err
		}

		err = tx.Create(user).Error
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

func UpdateUser(ctx context.Context, user *User) error {
	return dbutils.Transaction(func(tx *gorm.DB) error {
		password := ""
		if user.Password != "" {
			hashed, err := HashPassword(user.Password)
			if err != nil {
				return err
			}
			password = hashed
		}

		return tx.Exec(`
			UPDATE users
			SET user_name = ?, first_name = ?, last_name = ?, full_name = ?,
			    email = ?, phone_number = ?, is_admin = ?,
			    password_c = COALESCE(NULLIF(?, ''), password_c),
			    updated_at = NOW()
			WHERE id = ?
		`,
			user.UserName, user.FirstName, user.LastName, user.FullName,
			user.Email, user.PhoneNumber, user.IsAdmin,
			password, user.ID,
		).Error
	})
}

func DeleteUser(ctx context.Context, ids []int64) error {
	err := dbutils.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`DELETE FROM users WHERE id IN (?)`, ids).Error
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

func validateUser(ctx context.Context, user *User) error {
	return nil
}

func prepareData(ctx context.Context, user *User) error {
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return err
	}

	user.Password = hashedPassword

	return nil
}

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func getUser(ctx context.Context, username string) (hashedPassword string, found bool, err error) {
	db := dbutils.GetDB()

	user := User{}
	fetchResult := db.Model(User{}).Where(User{UserName: username}).Scan(&user)

	return user.Password, fetchResult.RowsAffected == 1, fetchResult.Error
}
