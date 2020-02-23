package models

import "az-fin/library/db"

type Contract struct {
	Model
	Date      string  `gorm:"type:varchar(10);unique_index:uidx_time_coin_hour;comment:'date'" json:"date"`
	CoinCapID string  `gorm:"type:varchar(255);unique_index:uidx_time_coin_hour;comment:'coin cap id'" json:"coin_cap_id"`
	BuyHour   int     `gorm:"unique_index:uidx_time_coin_hour;comment:'buy hour'" json:"buy_hour"`
	SaleHour  int     `gorm:"unique_index:uidx_time_coin_hour;comment:'sale hour'" json:"sale_hour"`
	Rate      float64 `gorm:"comment:'float base on buy hour'" json:"rate"`
}

func (c *Contract) Save() error {
	return db.DB.Save(c).Error
}
