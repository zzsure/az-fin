package cron

import (
	"az-fin/consts"
	"az-fin/library/util"
	"az-fin/library/util/net/http"
	"az-fin/models"
	"encoding/json"
	"github.com/op/go-logging"
	"github.com/robfig/cron"
	"github.com/buger/jsonparser"
)

var c *cron.Cron
var logger = logging.MustGetLogger("modules/cron")

type AssetResults []AssetResult
type AssetResult struct {
	CoinCapID         string `json:"id"`
	Rank              string `json:"rank"`
	Symbol            string `json:"symbol"`
	Name              string `json:"name"`
	Supply            string `json:"supply"`
	MaxSupply         string `json:"maxSupply"`
	MarketCapUsd      string `json:"marketCapUsd"`
	VolumeUsd24Hr     string `json:"volumeUsd24Hr"`
	PriceUsd          string `json:"priceUsd"`
	ChangePercent24Hr string `json:"changePercent24Hr"`
	Vwap24Hr          string `json:"vwap24Hr"`
}

func Init() {
	c = cron.New()
	getAssetsCron()
	c.Start()
}

func getAssetsCron() {
	//assetResults, err := getAssets()
	//if err != nil {
	//	logger.Error("get assets err: ", err)
	//}
	//dealAssetResults(assetResults)
	_ = c.AddFunc("@hourly", func() {
		assetResults, err := getAssets()
		if err != nil {
			logger.Error("get assets err: ", err)
		}
		dealAssetResults(assetResults)
	})
}

func dealAssetResults(assetResults AssetResults) {
	for _, assetResult := range assetResults {
		logger.Info("asset result is: ", assetResult)
		asset := &models.Asset{
			CoinCapID:         assetResult.CoinCapID,
			Rank:              assetResult.Rank,
			Symbol:            assetResult.Symbol,
			Name:              assetResult.Name,
			Supply:            assetResult.Supply,
			MaxSupply:         assetResult.MaxSupply,
			MarketCapUsd:      assetResult.MarketCapUsd,
			VolumeUsd24Hr:     assetResult.VolumeUsd24Hr,
			PriceUsd:          assetResult.PriceUsd,
			ChangePercent24Hr: assetResult.ChangePercent24Hr,
			Vwap24Hr:          assetResult.Vwap24Hr,
		}
		err := asset.Save()
		if err != nil {
			logger.Error("save asset to db err: ", err)
		}
	}
}

func getAssets() (AssetResults, error) {
	url := util.GetURL(consts.COINCAP_URL, consts.ASSETS_URI)
	b, err := http.Get(url, nil)
	if err != nil {
		return nil, err
	}
	data, _, _, _ := jsonparser.Get(b, "data")
	var assetResults AssetResults
	if err := json.Unmarshal(data, &assetResults); err != nil {
		return nil, err
	}
	return assetResults, nil
}