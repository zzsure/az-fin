package models

import "az-fin/library/db"

type Contract struct {
	Model
	Date         string  `gorm:"type:varchar(10);unique_index:uidx_date_coin_hour;comment:'date'" json:"date"`
	CoinCapID    string  `gorm:"type:varchar(255);unique_index:uidx_date_coin_hour;comment:'coin cap id'" json:"coin_cap_id"`
	BuyHour      int     `gorm:"unique_index:uidx_date_coin_hour;comment:'buy hour'" json:"buy_hour"`
	MaxSaleHour  int     `gorm:"unique_index:uidx_date_coin_hour;comment:'max sale hour'" json:"max_sale_hour"`
	MaxRate      float64 `gorm:"unique_index:uidx_date_coin_hour;comment:'max rate'" json:"max_rate"`
	BuyMillTime  int64   `gorm:"comment:'buy mill time'" json:"buy_mill_time"`
	SaleMillTime int64   `gorm:"comment:'sale mill time'" json:"sale_mill_time"`
	Rate         float64 `gorm:"comment:'float base on buy hour'" json:"rate"`
	BuyUsd       float64 `gorm:"comment:'buy usd'" json:"buy_usd"`
	SaleUsd      float64 `gorm:"comment:'sale usd'" json:"sale_usd"`
}

func (c *Contract) Save() error {
	return db.DB.Save(c).Error
}
