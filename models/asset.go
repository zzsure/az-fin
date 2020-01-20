package models

import "az-fin/library/db"

type Asset struct {
	Model
	CoinCapID         string `gorm:"type:varchar(255);unique_index:u_id_time_idx;comment:'unique identifier for asset'" json:"coin_cap_id"`
	Rank              string `gorm:"type:varchar(255);comment:'rank is in ascending order - this number is directly associated with the marketcap whereas the highest marketcap receives rank 1'" json:"rank"`
	Symbol            string `gorm:"type:varchar(255);comment:'most common symbol used to identify this asset on an exchange'" json:"symbol"`
	Name              string `gorm:"type:varchar(255);comment:'proper name for asset'" json:"name"`
	Supply            string `gorm:"type:varchar(255);comment:'available supply for trading'" json:"supply"`
	MaxSupply         string `gorm:"type:varchar(255);comment:'total quantity of asset issued'" json:"max_supply"`
	MarketCapUsd      string `gorm:"type:varchar(255);comment:'supply x price'" json:"market_cap_usd"`
	VolumeUsd24Hr     string `gorm:"type:varchar(255);comment:'quantity of trading volume represented in USD over the last 24 hours'" json:"volume_usd_24hr"`
	PriceUsd          string `gorm:"type:varchar(255);comment:'volume-weighted price based on real-time market data, translated to USD'" json:"price_usd"`
	ChangePercent24Hr string `gorm:"type:varchar(255);comment:'the direction and value change in the last 24 hours'" json:"change_percent_24hr"`
	Vwap24Hr          string `gorm:"type:varchar(255);comment:'Volume Weighted Average Price in the last 24 hours'" json:"vwap_24hr"`
	MillUnixTime      int64  `gorm:"unique_index:u_id_time_idx;comment:'timestamp in UNIX in milliseconds'" json:"mill_unix_time"`
}

func (a *Asset) Save() error {
	return db.DB.Save(a).Error
}
