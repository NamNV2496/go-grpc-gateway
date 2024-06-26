package domain

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username       string
	HashedPassword string
	Role           string
	IsActive       bool
}

func NewUser(
	user string,
	password string,
	role string,
	isActive bool,
) (*User, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &User{
		Username:       user,
		HashedPassword: string(hashPassword),
		Role:           role,
		IsActive:       isActive,
	}, nil
}

func (u User) TableName() string {
	return "user"
}

func GetUserByUserName(username string, db *gorm.DB) (User, error) {

	var user User
	err := db.Where("username=?", username).Find(&user).Error
	if err != nil {
		return User{}, err
	}
	return user, nil
}
