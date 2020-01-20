package asset

import (
	"az-fin/conf"
	"az-fin/consts"
	"az-fin/controller/response"
	"az-fin/library/redis"
	"az-fin/library/util"
	"az-fin/library/util/net/http"
	"az-fin/models"
	"encoding/json"
	"github.com/buger/jsonparser"
	"github.com/op/go-logging"
	"time"
)

var logger = logging.MustGetLogger("modules/asset")

func GetAssets() (response.AssetResults, int64, error) {
	url := util.GetURL(consts.COINCAP_URL, consts.ASSETS_URI)
	b, err := http.Get(url, nil)
	if err != nil {
		return nil, 0, err
	}
	data, _, _, _ := jsonparser.Get(b, "data")
	var assetResults response.AssetResults
	if err := json.Unmarshal(data, &assetResults); err != nil {
		return nil, 0, err
	}
	millUnixTime, _ := jsonparser.GetInt(b, "timestamp")

	return assetResults, millUnixTime, nil
}

func DealAssetResults(assetResults response.AssetResults, millUnixTime int64) ([]*models.Asset, error) {
	assets := make([]*models.Asset, 0)
	for _, assetResult := range assetResults {
		//logger.Info("asset result is: ", assetResult)
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
			MillUnixTime:      millUnixTime,
		}
		err := asset.Save()
		if err != nil {
			logger.Error("save asset to db err: ", err)
			continue
		}
		assets = append(assets, asset)
	}
	if conf.Config.Redis.IsUse {
		b, _ := json.Marshal(assets)
		redis.GoRedisClient.Set(consts.COINCAP_ASSETS_KEY, string(b), time.Second*consts.REDIS_KEY_EXPIRED_SECONDS)
	}
	return assets, nil
}
