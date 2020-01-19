package cron

import (
	"az-fin/consts"
	"az-fin/library/util"
	"az-fin/library/util/net/http"
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
	getAssets()
	c.AddFunc("@hourly", func() {
		getAssets()
	})
}

func getAssets() {
	url := util.GetURL(consts.COINCAP_URL, consts.ASSETS_URI)
	b, err := http.Get(url, nil)
	if err != nil {
		logger.Error("get coincap assets err:", err)
	}
	logger.Info(string(b))
}
