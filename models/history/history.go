package history

import (
	"../../db"
	"github.com/jinzhu/gorm"
)

type Notation struct {
	gorm.Model

	Number int
	Img    string
	UserID int64 `gorm:"not null;index"`
}

func (n *Notation) Save() error {
	return db.Get().Create(n).Error
}

func init() {
	//	if !cfg.IsProduction() {
	db.Get().AutoMigrate(&Notation{})
	//	}
}
