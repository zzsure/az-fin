package models

import "az-fin/library/db"

type Contract struct {
	Model
	Date         string  `gorm:"type:varchar(10);comment:'date'" json:"date"`
	CoinCapID    string  `gorm:"type:varchar(255);comment:'coin cap id'" json:"coin_cap_id"`
	BuyHour      int     `gorm:"comment:'buy hour'" json:"buy_hour"`
	MaxSaleHour  int     `gorm:"comment:'max sale hour'" json:"max_sale_hour"`
	MaxRate      float64 `gorm:"comment:'max rate'" json:"max_rate"`
	SaleMillTime int64   `gorm:"comment:'sale mill time'" json:"sale_mill_time"`
	Rate         float64 `gorm:"comment:'float base on buy hour'" json:"rate"`
}

func (c *Contract) Save() error {
	return db.DB.Save(c).Error
}
