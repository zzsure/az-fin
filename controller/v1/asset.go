package v1

import (
	"az-fin/consts"
	"az-fin/controller/response"
	"az-fin/library/util"
	"az-fin/library/util/net/http"
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/gin-gonic/gin"
	"strconv"
)

func PriceExcel(c *gin.Context) {
	coinCapIDs := [10]string{"bitcoin", "ethereum", "ripple", "bitcoin-cash", "litecoin", "binance-coin", "eos", "bitcoin-sv", "monero", "huobi-token"}
	prices := [10]float64{}
	var err error
	for idx, id := range coinCapIDs {
		prices[idx], err = getCoinCapPrice(id)
		if err != nil {
			response.ServerLogErr(c, logger, "get coincap price not float: "+err.Error())
			return
		}
	}
	priceText := ""
	for _, price := range prices {
		priceText += fmt.Sprintf("%f", price) + "\n"
	}
	rmbRate, err := getRate(consts.RMB_COINCAP_ID)
	if err != nil {
		response.ServerLogErr(c, logger, "get coincap price not float: "+err.Error())
		return
	}
	priceText += fmt.Sprintf("%f", rmbRate)
	c.String(200, priceText)
	//response.ServerSucc(c, "success", prices)
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
	url = fmt.Sprintf("%s/%s", url, id)
	b, err := http.Get(url, nil)
	if err != nil {
		return 0.0, err
	}
	data, _, _, _ := jsonparser.Get(b, "data")
	var assetResult response.AssetResult
	if err := json.Unmarshal(data, &assetResult); err != nil {
		return 0.0, err
	}
	return strconv.ParseFloat(assetResult.PriceUsd, 64)
}

func PriceTest(c *gin.Context) {
	prices := [11]float64{0.59, 100.9, 10.9, 8.9, 10.2, 11, 12.9}
	priceText := ""
	for _, price := range prices {
		priceText += fmt.Sprintf("%f", price) + "\n"
	}
	c.String(200, priceText)
	//response.ServerSucc(c, "success", priceText)
}