package models

import "az-fin/library/db"

type ContractOrder struct {
	Model
	Date         string  `gorm:"type:varchar(10);comment:'date'" json:"date"`
	Symbol       string  `gorm:"type:varchar(255);comment:'symbol'" json:"symbol"`
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
	Status       int     `gorm:"comment:'order status, 1: buy, 2: sale'" json:"status"`
	RandomHour   int     `gorm:"comment:'random hour'" json:"random_hour"`
}

func (co *ContractOrder) Save() error {
	return db.DB.Save(co).Error
}

func GetLastContractOrder(symbol string) (*ContractOrder, error) {
	var co ContractOrder
	err := db.DB.Order("buy_mill_time desc").Where("symbol = ?", symbol).Last(&co).Error
	return &co, err
}

func GetAllContractOrders(symbol string) ([]*ContractOrder, error) {
	var cos []*ContractOrder
	err := db.DB.Order("buy_mill_time asc").Where("symbol = ?", symbol).Find(&cos).Error
	return cos, err
}

func GetContractOrdersByBatchID(symbol string, id int) ([]*ContractOrder, error) {
	var cos []*ContractOrder
	err := db.DB.Where("symbol = ? and batch_id = ?", symbol, id).Find(&cos).Error
	return cos, err
}
