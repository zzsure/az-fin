package tool

import (
	"az-fin/conf"
	"az-fin/library/db"
	"az-fin/library/log"
	"az-fin/library/util"
	"az-fin/models"
	"github.com/urfave/cli"
)

// 获取各种币t1-t2历史上x点开仓，y点平仓，上下浮动分别的次数

var Contract = cli.Command{
	Name:  "contract",
	Usage: "az-fin contract data",
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
	Action: runContract,
}

func runContract(c *cli.Context) {
	conf.Init(c.String("conf"), c.String("args"))
	log.Init()
	db.Init()
	db.DB.LogMode(conf.Config.Database.LogMode)
	if conf.Config.Contract.EndMillTime <= conf.Config.Contract.StartMillTime {
		logger.Error("end mill time should greater than start mill time")
		return
	}
	if conf.Config.Contract.SaleHour <= conf.Config.Contract.BuyHour {
		logger.Error("sale hour should greater than buy hour")
		return
	}

	t := util.GetTimeByMillUnixTime(conf.Config.Contract.StartMillTime)
	m := util.GetMorningUnixTime(t)
	st := (m + int64(conf.Config.Contract.BuyHour)*60*60) * 1000

	for ; st < conf.Config.Contract.EndMillTime; st += 24 * 60 * 60 * 1000 {
		t = util.GetTimeByMillUnixTime(st)
		date := util.GetDateByTime(t)
		et := st + int64(conf.Config.Contract.SaleHour-conf.Config.Contract.BuyHour)*60*60*1000
		sp, err := models.GetPricesByMillTime(conf.Config.Contract.CoinCapID, st)
		if err != nil {
			continue
		}
		ep, err := models.GetPricesByMillTime(conf.Config.Contract.CoinCapID, et)
		rate := 0.0
		if sp.PriceUsd != 0.0 {
			rate = ep.PriceUsd/sp.PriceUsd - 1.0
		}
		contract := &models.Contract{
			Date:      date,
			CoinCapID: conf.Config.Contract.CoinCapID,
			BuyHour:   conf.Config.Contract.BuyHour,
			SaleHour:  conf.Config.Contract.SaleHour,
			Rate:      rate,
		}
		err = contract.Save()
		if err != nil {
			logger.Error("save contract err: ", err)
		}
	}
}
