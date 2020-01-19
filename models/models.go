package models

import (
	"az-fin/library/db"
	"github.com/op/go-logging"
	"time"
)

type Model struct {
	ID        uint      `gorm:"primary_key;auto_increment" json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

var logger = logging.MustGetLogger("models")

func CreateTable() {
	db.DB.DropTableIfExists(&Asset{})
	db.DB.DropTableIfExists(&Price{})
	create := db.DB.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8")
	create.CreateTable(&Asset{})
	create.CreateTable(&Price{})
}

func MigrateTable() {
	create := db.DB.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8")
	create.AutoMigrate(&Asset{})
	create.AutoMigrate(&Price{})
}
