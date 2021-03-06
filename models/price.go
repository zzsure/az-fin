package models

import "az-fin/library/db"

type Price struct {
	Model
	Symbol            string  `gorm:"type:varchar(255);unique_index:u_id_time_idx;comment:'unique identifier for asset'" json:"symbol"`
	CoinCapID         string  `gorm:"type:varchar(255);unique_index:u_id_time_idx;comment:'unique identifier for asset'" json:"coin_cap_id"`
	Interval          string  `gorm:"type:varchar(255);comment:'point-in-time interval. minute and hour intervals represent price at that time, the day interval represents average of 24 hour periods (timezone: UTC)'" json:"interval"`
	PriceUsd          float64 `gorm:"comment:'volume-weighted price based on real-time market data, translated to USD'" json:"price_usd"`
	CirculatingSupply float64 `gorm:"default:0.0;comment:'circulating supply'" json:"circulating supply"`
	MillUnixTime      int64   `gorm:"unique_index:u_id_time_idx;comment:'timestamp in UNIX in milliseconds'" json:"mill_unix_time"`
}

func (p *Price) Save() error {
	return db.DB.Save(p).Error
}

// TODO: bash insert
func BashInsertPrice(prices []*Price) error {
	return nil
}

func GetPriceByMillTime(id string, mt int64) (*Price, error) {
	var p Price
	err := db.DB.Where("symbol = ? and mill_unix_time = ?", id, mt).Find(&p).Error
	return &p, err
}

func GetPricesBySymbolAndTime(symbol string, mt1, mt2 int64) ([]*Price, error) {
	var ps []*Price
	err := db.DB.Order("mill_unix_time asc").Where("symbol = ? and mill_unix_time >= ? and mill_unix_time < ?", symbol, mt1, mt2).Find(&ps).Error
	return ps, err
}
