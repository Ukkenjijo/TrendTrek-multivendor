package models

import "gorm.io/gorm"

type Role string



type User struct {
	gorm.Model
	ID             uint   `json:"id" gorm:"primaryKey;type:serial"`
	Name           string `gorm:"column:name;type:varchar(255)" validate:"required" json:"name"`
	Email          string `json:"email" gorm:"unique;not null"`
	PhoneNumber    string `json:"phone" gorm:"unique;not null"`
	Blocked        bool   `gorm:"column:blocked;type:bool" json:"blocked"`
	Role           int     `json:"role" gorm:"default:1"`
	HashedPassword string `gorm:"column:hashed_password;type:varchar(255)" validate:"required,min=8" json:"hashed_password"`
	Verified       bool   `json:"verified" gorm:"default:false"`
	IsAdmin        bool   `gorm:"default:false"`
	ProfilePicture string `gorm:"size:255"`
}
