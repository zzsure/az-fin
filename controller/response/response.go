package response

import (
	"az-fin/consts"
	"az-fin/library/log"
	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
)

func Response(c *gin.Context, code int, msg string, data interface{}) {
	requestId := c.MustGet(consts.REQUEST_ID_KEY)
	c.JSON(consts.HTTP_SUCC_CODE, map[string]interface{}{
		"data":       data,
		"error_no":   code,
		"error_msg":  msg,
		"request_id": requestId,
	})
}

func ServerSucc(c *gin.Context, msg string, data interface{}) {
	requestId := c.MustGet(consts.REQUEST_ID_KEY).(log.RequestID)
	c.JSON(consts.HTTP_SUCC_CODE, map[string]interface{}{
		"data":       data,
		"error_no":   consts.ERR_NO_SUCC,
		"error_msg":  msg,
		"request_id": requestId,
	})
}

func ClientErr(c *gin.Context, msg string) {
	Response(c, consts.ERR_NO_CLIENT_COMMON, msg, nil)
}

func ClientNoErr(c *gin.Context, errNo int) {
	Response(c, errNo, consts.GetErrMsg(errNo), nil)
}

func ServerErr(c *gin.Context, msg string) {
	Response(c, consts.ERR_NO_SYSTEM_COMMON, msg, nil)
}

func ServerNoErr(c *gin.Context, errNo int) {
	Response(c, errNo, consts.GetErrMsg(errNo), nil)
}

func ServerLogErr(c *gin.Context, logger *logging.Logger, msg string) {
	requestId := c.MustGet(consts.REQUEST_ID_KEY).(log.RequestID)
	logger.Error(requestId, msg)
	ServerErr(c, msg)
}

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
type AssetResults []AssetResult

type RateResult struct {
	CoinCapID      string `json:"id"`
	Symbol         string `json:"symbol"`
	CurrencySymbol string `json:"currencySymbol"`
	RateUsd        string `json:"rateUsd"`
	Type           string `json:"type"`
}
type RateResults []RateResult

type PriceResult struct {
	PriceUsd          string `json:"priceUsd"`
	Time              int64  `json:"time"`
	CirculatingSupply string `json:"circulatingSupply"`
}
type PriceResults []PriceResult
