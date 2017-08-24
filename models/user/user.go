package user

import (

	"github.com/jinzhu/gorm"
	"../../db"
)

type User struct {
	gorm.Model

	Username  string `gorm:"not null;unique_index"`
	Password  string
}

func (u *User) Get() error {
	return db.Get().First(u).Error
}

func (u *User) Create() error {
	return db.Get().Create(u).Error
}

func init() {
	db.Get().AutoMigrate(&User{})
}
