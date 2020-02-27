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
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/op/go-logging"
	"strconv"
	"strings"
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

func GetPrices(coinCapID, interval string, start, end int64) (response.PriceResults, error) {
	url := fmt.Sprintf("%s/%s/%s/history?interval=%s&start=%d&end=%d", strings.TrimRight(consts.COINCAP_URL, "/"), strings.TrimLeft(consts.ASSETS_URI, "/"), coinCapID, interval, start, end)
	logger.Info("get url: ", url)
	b, err := http.Get(url, nil)
	if err != nil {
		return nil, err
	}
	data, _, _, _ := jsonparser.Get(b, "data")
	var priceResults response.PriceResults
	if err := json.Unmarshal(data, &priceResults); err != nil {
		return nil, err
	}
	logger.Info("get price len: ", len(priceResults))
	return priceResults, nil
}

func DealPrices(priceResults response.PriceResults) error {
	for _, priceResult := range priceResults {
		p, err := strconv.ParseFloat(priceResult.PriceUsd, 64)
		if err != nil {
			return err
		}
		s, _ := strconv.ParseFloat(priceResult.CirculatingSupply, 64)
		price := &models.Price{
			CoinCapID:         conf.Config.History.CoinCapID,
			Symbol:            conf.Config.History.Symbol,
			Interval:          conf.Config.History.Interval,
			PriceUsd:          p,
			MillUnixTime:      priceResult.Time,
			CirculatingSupply: s,
		}
		err = price.Save()
		if err != nil {
			return err
		}
	}
	return nil
}
