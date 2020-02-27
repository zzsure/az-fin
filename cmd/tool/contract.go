package tool

import (
	"az-fin/conf"
	"az-fin/library/db"
	"az-fin/library/log"
	"az-fin/library/util"
	"az-fin/models"
	"github.com/urfave/cli"
)

// 获取各种币t1-t2历史上x点开仓，y点平仓，上下浮动分别的次数 git log: 60139e7cc15e8fc69f97b545f5588130af25cfa5
// 根据买入时间，买入浮动上下波动范围后平仓，最晚时间平仓计算每日数据

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
	if conf.Config.Contract.MaxSaleHour <= conf.Config.Contract.BuyHour {
		logger.Error("sale hour should greater than buy hour")
		return
	}

	t := util.GetTimeByMillUnixTime(conf.Config.Contract.StartMillTime)
	m := util.GetMorningUnixTime(t)
	st := (m + int64(conf.Config.Contract.BuyHour)*60*60) * 1000

	for ; st < conf.Config.Contract.EndMillTime; st += 24 * 60 * 60 * 1000 {
		t = util.GetTimeByMillUnixTime(st)
		date := util.GetDateByTime(t)
		logger.Info("deal date: ", date)

		sp, err := models.GetPriceByMillTime(conf.Config.Contract.CoinCapID, st)
		if err != nil {
			continue
		}

		et := st + int64(conf.Config.Contract.MaxSaleHour-conf.Config.Contract.BuyHour)*60*60*1000
		ep, err := models.GetPriceByMillTime(conf.Config.Contract.CoinCapID, et)
		if err != nil {
			continue
		}
		smt := et

		for it := st; it <= et; it += 60 * 1000 {
			cp, err := models.GetPriceByMillTime(conf.Config.Contract.CoinCapID, it)
			if err != nil {
				continue
			}
			// 上涨止损
			if cp.PriceUsd >= (1+conf.Config.Contract.MaxRate)*sp.PriceUsd {
				smt = cp.MillUnixTime
				ep = cp
				break
			} else if cp.PriceUsd <= (1-conf.Config.Contract.MaxRate)*sp.PriceUsd {
				smt = cp.MillUnixTime
				ep = cp
				break
			}
		}

		rate := 0.0
		if sp.PriceUsd != 0.0 {
			rate = ep.PriceUsd/sp.PriceUsd - 1.0
		} else {
			continue
		}
		contract := &models.Contract{
			Date:         date,
			CoinCapID:    conf.Config.Contract.CoinCapID,
			BuyHour:      conf.Config.Contract.BuyHour,
			MaxSaleHour:  conf.Config.Contract.MaxSaleHour,
			MaxRate:      conf.Config.Contract.MaxRate,
			BuyMillTime:  st,
			SaleMillTime: smt,
			Rate:         rate,
			BuyUsd:       sp.PriceUsd,
			SaleUsd:      ep.PriceUsd,
		}
		err = contract.Save()
		if err != nil {
			logger.Error("save contract err: ", err)
		}
	}
}
