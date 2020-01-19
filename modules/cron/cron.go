package cron

import (
	"az-fin/consts"
	"az-fin/controller/response"
	"az-fin/library/util"
	"az-fin/library/util/net/http"
	"az-fin/models"
	"encoding/json"
	"github.com/buger/jsonparser"
	"github.com/op/go-logging"
	"github.com/robfig/cron"
)

var c *cron.Cron
var logger = logging.MustGetLogger("modules/cron")

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

func dealAssetResults(assetResults response.AssetResults) {
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
			MillUnixTime:      util.GetMillUnixTime(),
		}
		err := asset.Save()
		if err != nil {
			logger.Error("save asset to db err: ", err)
		}
	}
}

func getAssets() (response.AssetResults, error) {
	url := util.GetURL(consts.COINCAP_URL, consts.ASSETS_URI)
	b, err := http.Get(url, nil)
	if err != nil {
		return nil, err
	}
	data, _, _, _ := jsonparser.Get(b, "data")
	var assetResults response.AssetResults
	if err := json.Unmarshal(data, &assetResults); err != nil {
		return nil, err
	}
	return assetResults, nil
}
