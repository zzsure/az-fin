package models

type Asset struct {
	Model
	CoinCapID         string `gorm:"type:varchar(255);comment:'unique identifier for asset'" json:"id"`
	Rank              string `gorm:"type:varchar(255);comment:'rank is in ascending order - this number is directly associated with the marketcap whereas the highest marketcap receives rank 1'" json:"rank"`
	Symbol            string `gorm:"type:varchar(255);comment:'most common symbol used to identify this asset on an exchange'" json:"symbol"`
	Name              string `gorm:"type:varchar(255);comment:'proper name for asset'" json:"name"`
	Supply            string `gorm:"type:varchar(255);comment:'available supply for trading'" json:"supply"`
	MaxSupply         string `gorm:"type:varchar(255);comment:'total quantity of asset issued'" json:"maxSupply"`
	MarketCapUsd      string `gorm:"type:varchar(255);comment:'supply x price'" json:"marketCapUsd"`
	VolumeUsd24Hr     string `gorm:"type:varchar(255);comment:'quantity of trading volume represented in USD over the last 24 hours'" json:"volumeUsd24Hr"`
	PriceUsd          string `gorm:"type:varchar(255);comment:'volume-weighted price based on real-time market data, translated to USD'" json:"priceUsd"`
	ChangePercent24Hr string `gorm:"type:varchar(255);comment:'the direction and value change in the last 24 hours'" json:"changePercent24Hr"`
	Vwap24Hr          string `gorm:"type:varchar(255);comment:'Volume Weighted Average Price in the last 24 hours'" json:"vwap24Hr"`
}
