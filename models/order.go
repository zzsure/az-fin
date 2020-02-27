package models

import "az-fin/library/db"

type ContractOrder struct {
	Model
	Date         string  `gorm:"type:varchar(10);unique_index:uidx_date_coin_hour;comment:'date'" json:"date"`
	Symbol       string  `gorm:"type:varchar(255);unique_index:uidx_date_coin_hour;comment:'symbol'" json:"symbol"`
	BatchID      int     `gorm:"comment:'batch id'" json:"batch_id"`
	Depth        int     `gorm:"comment:'depth'" json:"depth"`
	StartBalance float64 `gorm:"comment:'start coin balance'" json:"start_balance"`
	EndBalance   float64 `gorm:"comment:'end coin balance'" json:"end_balance"`
	CoinAmount   float64 `gorm:"comment:'buy coin amount'" json:"coin_amount"`
	ContractNum  int     `gorm:"comment:'contract num'" json:"contract_num"`
	BuyPrice     float64 `gorm:"comment:'buy price'" json:"buy_price"`
	SalePrice    float64 `gorm:"comment:'sale price'" json:"sale_sale"`
	BuyUsd       float64 `gorm:"comment:'buy usd'" json:"buy_usd"`
	BuyMillTime  int64   `gorm:"comment:'buy mill time'" json:"buy_mill_time"`
	SaleMillTime int64   `gorm:"comment:'sale mill time'" json:"sale_mill_time"`
	MaxRate      float64 `gorm:"comment:'max rate'" json:"max rate"`
	Rate         float64 `gorm:"comment:'float base on buy hour'" json:"rate"`
	Fee          float64 `gorm:"comment:'coin fee'" json:"fee"`
	Profit       float64 `gorm:"comment:'coin profit'" json:"profit"`
}

func (co *ContractOrder) Save() error {
	return db.DB.Save(co).Error
}

func GetLastContractOrder(symbol string) (*ContractOrder, error) {
	var co ContractOrder
	err := db.DB.Order("buy_mill_time desc").Where("symbol = ?", symbol).Last(&co).Error
	return &co, err
}
