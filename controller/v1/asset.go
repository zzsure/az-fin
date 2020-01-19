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
	coinCapIDs := [11]string{"bitcoin", "ethereum", "ripple", "bitcoin-cash", "litecoin", "binance-coin", "eos", "bitcoin-sv", "monero", "huobi-token", "tether"}
	prices := [11]float64{}
	for idx, id := range coinCapIDs {
		url := util.GetURL(consts.COINCAP_URL, consts.ASSETS_URI)
		url = fmt.Sprintf("%s/%s", url, id)
		b, err := http.Get(url, nil)
		if err != nil {
			response.ServerLogErr(c, logger, "get price err: "+err.Error())
		}
		data, _, _, _ := jsonparser.Get(b, "data")
		var assetResult response.AssetResult
		if err := json.Unmarshal(data, &assetResult); err != nil {
			response.ServerLogErr(c, logger, "json unmarshal err: "+err.Error())
		}
		prices[idx], err = strconv.ParseFloat(assetResult.PriceUsd, 64)
		if err != nil {
			response.ServerLogErr(c, logger, "get price not float: "+err.Error())
		}
	}
	priceText := ""
	for _, price := range prices {
		priceText += fmt.Sprintf("%f", price) + "\n"
	}
	c.String(200, priceText)
	//response.ServerSucc(c, "success", prices)
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
