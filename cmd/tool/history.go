package tool

import (
	"az-fin/conf"
	"az-fin/library/db"
	"az-fin/library/log"
	"az-fin/library/util"
	"az-fin/modules/asset"
	"github.com/op/go-logging"
	"github.com/urfave/cli"
)

var logger = logging.MustGetLogger("cmd/tool")

// 根据时间范围获取对应的prices的数据

var History = cli.Command{
	Name:  "history",
	Usage: "az-fin history data",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "conf, c",
			Value: "config.toml",
			Usage: "toml配置文件",
		},
		cli.StringFlag{
			Name:  "args",
			Value: "",
			Usage: "multi config cmd line args",
		},
	},
	Action: runHistory,
}

func runHistory(c *cli.Context) {
	conf.Init(c.String("conf"), c.String("args"))
	log.Init()
	db.Init()
	db.DB.LogMode(conf.Config.Database.LogMode)
	if conf.Config.History.EndMillTime <= conf.Config.History.StartMillTime {
		logger.Error("end mill time should greater than start mill time")
		return
	}
	for t := conf.Config.History.StartMillTime; t < conf.Config.History.EndMillTime; t += 24 * 60 * 60 * 1000 {
		time := util.GetTimeByMillUnixTime(t)
		logger.Info("deal data time: ", util.GetFormatTime(time))
		priceResults, err := asset.GetPrices(conf.Config.History.CoinCapID, conf.Config.History.Interval, t, t+24*60*60*1000)
		if err != nil {
			logger.Error("get price result err: ", err)
			break
		}
		err = asset.DealPrices(priceResults)
		if err != nil {
			logger.Error("save price err: ", err)
			continue
		}
		util.RandomSleep(conf.Config.History.MaxSleepSecond)
	}
}
