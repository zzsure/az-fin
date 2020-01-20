package cron

import (
	"az-fin/modules/asset"

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
	// hourly,daily,weekly,monthly,yearly
	_ = c.AddFunc("every 1m", func() {
		assetResults, millUnixTime, err := asset.GetAssets()
		if err != nil {
			logger.Error("get assets err: ", err)
		}
		asset.DealAssetResults(assetResults, millUnixTime)
	})
}
