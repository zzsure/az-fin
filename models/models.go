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
	db.DB.DropTableIfExists(&Contract{})
	db.DB.DropTableIfExists(&ContractOrder{})
	db.DB.DropTableIfExists(&Order{})
	db.DB.DropTableIfExists(&Profit{})
	create := db.DB.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8")
	create.CreateTable(&Asset{})
	create.CreateTable(&Price{})
	create.CreateTable(&Contract{})
	create.CreateTable(&ContractOrder{})
	create.CreateTable(&Order{})
	create.CreateTable(&Profit{})
}

func MigrateTable() {
	create := db.DB.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8")
	create.AutoMigrate(&Asset{})
	create.AutoMigrate(&Price{})
	create.AutoMigrate(&Contract{})
	create.AutoMigrate(&ContractOrder{})
	create.AutoMigrate(&Order{})
	create.AutoMigrate(&Profit{})
}
