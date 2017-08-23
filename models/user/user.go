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

func (u *User) GetOrCreate() error {
	return db.Get().FirstOrCreate(u, User{Username: u.Username, Password: u.Password}).Error
}

func init() {
	db.Get().AutoMigrate(&User{})
}
