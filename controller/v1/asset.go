package v1

import (
	"az-fin/conf"
	"az-fin/consts"
	"az-fin/controller/response"
	"az-fin/library/redis"
	"az-fin/library/util"
	"az-fin/library/util/net/http"
	"az-fin/models"
	"az-fin/modules/asset"
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/gin-gonic/gin"
	goRedis "github.com/go-redis/redis"
	"strconv"
	"time"
)

func PriceExcel(c *gin.Context) {
	const tryTimes = 20
	const NUM = 17
	coinCapIDs := [NUM]string{"bitcoin", "ethereum", "xrp", "bitcoin-cash", "litecoin", "binance-coin", "eos",
		"bitcoin-sv", "monero", "huobi-token", "ethereum-classic", "dash", "zcash", "chainlink", "polkadot", "tron", "yearn-finance"}
	prices := [NUM]float64{}
	var err error
	for idx, id := range coinCapIDs {
		for tryTime := 0; tryTime < tryTimes; tryTime++ {
			prices[idx], err = getCoinCapPrice(id)
			if err == nil {
				break
			}
			time.Sleep(time.Duration(1) * time.Second)
			tryTime++
		}
		if err != nil {
			response.ServerLogErr(c, logger, "get coincap price: "+err.Error())
			return
		}
	}
	priceText := ""
	for _, price := range prices {
		priceText += fmt.Sprintf("%f", price) + "\n"
	}
	rmbRate, err := getRate(consts.RMB_COINCAP_ID)
	if err != nil {
		response.ServerLogErr(c, logger, "get coincap rate: "+err.Error())
		return
	}
	priceText += fmt.Sprintf("%f", rmbRate)

	huobiBalance := 0.0
	if conf.Config.Redis.IsUse {
		cacheData, err := redis.GoRedisClient.Get(consts.HUOBI_BALANCE_KEY).Result()
		if err != goRedis.Nil {
			balanceMap := make(map[string]float64)
			err := json.Unmarshal([]byte(cacheData), &balanceMap)
			if err == nil {
				usdt := 0.0
				btc := 0.0
				if _, ok := balanceMap["usdt"]; ok {
					usdt = balanceMap["usdt"]
				}
				if _, ok := balanceMap["btc"]; ok {
					btc = balanceMap["btc"]
				}
				huobiBalance += btc * (1 / rmbRate) * prices[0]
				huobiBalance += usdt * (1 / rmbRate)
			}
		}
	}

	priceText += "\n" + "\n" + fmt.Sprintf("%f", huobiBalance)

	c.String(200, priceText)
	//response.ServerSucc(c, "success", prices)
}

func PriceList(c *gin.Context) {
	results := make([]*models.Asset, 0)
	isNeedRequest := false
	if conf.Config.Redis.IsUse {
		cacheData, err := redis.GoRedisClient.Get(consts.COINCAP_ASSETS_KEY).Result()
		// It returns redis.Nil error when key does not exist.
		if err == goRedis.Nil {
			isNeedRequest = true
		} else {
			logger.Info("get asset by redis")
			err := json.Unmarshal([]byte(cacheData), &results)
			if err != nil {
				response.ServerLogErr(c, logger, "json unmarshal asset err: "+err.Error())
				return
			}
		}
	} else {
		isNeedRequest = true
	}
	if isNeedRequest == true {
		logger.Info("get asset by request")
		assetResults, millUnixTime, err := asset.GetAssets()
		if err != nil {
			response.ServerLogErr(c, logger, "get coincap asset err: "+err.Error())
			return
		}
		results, err = asset.DealAssetResults(assetResults, millUnixTime)
		if err != nil {
			response.ServerLogErr(c, logger, "deal coincap asset err: "+err.Error())
			return
		}
	}
	response.ServerSucc(c, "get assets success", results)
}

func getRate(id string) (float64, error) {
	url := util.GetURL(consts.COINCAP_URL, consts.RATES_URI)
	b, err := http.Get(url, nil)
	if err != nil {
		return 0.0, err
	}
	data, _, _, _ := jsonparser.Get(b, "data")
	var rateResults response.RateResults
	if err := json.Unmarshal(data, &rateResults); err != nil {
		return 0.0, err
	}
	rateUSD := ""
	for _, rateResult := range rateResults {
		if rateResult.CoinCapID == id {
			rateUSD = rateResult.RateUsd
			break
		}
	}
	return strconv.ParseFloat(rateUSD, 64)
}

func getCoinCapPrice(id string) (float64, error) {
	url := util.GetURL(consts.COINCAP_URL, consts.ASSETS_URI)
	// TODO:优化多个一起查询
	url = fmt.Sprintf("%s/%s", url, id)
	logger.Info("url is: %s", url)
	b, err := http.Get(url, nil)
	if err != nil {
		return 0.0, err
	}
	data, _, _, err := jsonparser.Get(b, "data")
	if nil != err {
		return 0.0, err
	}
	var assetResult response.AssetResult
	if err := json.Unmarshal(data, &assetResult); err != nil {
		return 0.0, err
	}
	return strconv.ParseFloat(assetResult.PriceUsd, 64)
}
