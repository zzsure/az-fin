package tool

import (
	"az-fin/cmd"
	"az-fin/conf"
	"az-fin/consts"
	"az-fin/library/db"
	"az-fin/library/log"
	"az-fin/library/util"
	"az-fin/models"
	"github.com/urfave/cli"
)

// 1: 分析初始合约x张，最大接盘金额，盈利数据

var Analyze = cli.Command{
	Name:  "analyze",
	Usage: "az-fin analyze data",
	Flags: []cli.Flag{
		cmd.StringFlag("conf, c", "config.toml", "toml配置文件"),
		cmd.StringFlag("args, a", "", "cmd line args"),
		cmd.IntFlag("type", 0, "api type"),
	},
	Action: runAnalyze,
}

func runAnalyze(c *cli.Context) {
	conf.Init(c.String("conf"), c.String("args"))
	log.Init()
	db.Init()
	db.DB.LogMode(conf.Config.Database.LogMode)
	t := c.Int("type")
	switch t {
	case consts.CONTRACT_ORDER_FIX_BUY_HOUR:
		fixBuyHour()
	case consts.CONTRACT_ORDER_RANDOM_GAP:
		randomGap()
	}
}

func randomGap() {
	if conf.Config.Analyze.EndMillTime <= conf.Config.Analyze.StartMillTime {
		logger.Error("end mill time should greater than start mill time")
		return
	}
	prices, err := models.GetPricesBySymbolAndTime(conf.Config.Analyze.Symbol, conf.Config.Analyze.StartMillTime, conf.Config.Analyze.EndMillTime)
	if err != nil {
		logger.Error("get prices " + err.Error())
		return
	}
	logger.Info("get prices len: ", len(prices))
	priceMap := make(map[int64]*models.Price, len(prices))
	for _, p := range prices {
		priceMap[p.MillUnixTime] = p
	}

	for st := conf.Config.Analyze.StartMillTime; st < conf.Config.Analyze.EndMillTime; st += 60 * 1000 {
		t := util.GetTimeByMillUnixTime(st)
		date := util.GetDateByTime(t)
		logger.Info("deal date: ", date)

	}
}

func fixBuyHour() {
	if conf.Config.Analyze.EndMillTime <= conf.Config.Analyze.StartMillTime {
		logger.Error("end mill time should greater than start mill time")
		return
	}

	if conf.Config.Analyze.MaxSaleHour <= conf.Config.Analyze.BuyHour {
		logger.Error("sale hour should greater than buy hour")
		return
	}

	prices, err := models.GetPricesBySymbolAndTime(conf.Config.Analyze.Symbol, conf.Config.Analyze.StartMillTime, conf.Config.Analyze.EndMillTime)
	if err != nil {
		logger.Error("get prices " + err.Error())
		return
	}
	logger.Info("get prices len: ", len(prices))
	priceMap := make(map[int64]*models.Price, len(prices))
	for _, p := range prices {
		priceMap[p.MillUnixTime] = p
	}

	t := util.GetTimeByMillUnixTime(conf.Config.Analyze.StartMillTime)
	m := util.GetMorningUnixTime(t)
	st := (m + int64(conf.Config.Analyze.BuyHour)*60*60) * 1000

	for ; st < conf.Config.Analyze.EndMillTime; st += 24 * 60 * 60 * 1000 {
		t = util.GetTimeByMillUnixTime(st)
		date := util.GetDateByTime(t)
		logger.Info("deal date: ", date)

		sp, ok := priceMap[st]
		if !ok {
			logger.Error("not have price: ", st)
			continue
		}
		et := st + int64(conf.Config.Analyze.MaxSaleHour-conf.Config.Analyze.BuyHour)*60*60*1000
		ep, ok := priceMap[et]
		if !ok {
			logger.Error("not have price: ", et)
			continue
		}
		epUsd := ep.PriceUsd
		smt := et

		contractNum := conf.Config.Analyze.InitContractNum
		lastBalance := 0.0
		buyAmount := 0.0
		buyUsd := 0.0
		batchID := 1
		depth := 1
		fee := 0.0

		co, err := models.GetLastContractOrder(conf.Config.Analyze.Symbol)
		if err == nil {
			// 上次交易亏损
			if co.Profit < 0.0 {
				contractNum = 2 * co.ContractNum
				batchID = co.BatchID
				depth = co.Depth + 1
			} else {
				batchID = co.BatchID + 1
			}
			lastBalance = co.EndBalance
		}

		contractUsd := 10.0 * float64(contractNum/20)
		coinAmount := contractUsd / sp.PriceUsd
		if lastBalance*sp.PriceUsd < contractUsd {
			buyUsd = contractUsd - lastBalance*sp.PriceUsd
			buyAmount = buyUsd / sp.PriceUsd
		}

		for it := st; it <= et; it += 60 * 1000 {
			cp, ok := priceMap[st]
			if !ok {
				continue
			}
			// 上涨止损
			if cp.PriceUsd >= (1+conf.Config.Analyze.MaxRate)*sp.PriceUsd {
				smt = cp.MillUnixTime
				epUsd = (1 + conf.Config.Analyze.MaxRate) * sp.PriceUsd
				break
			} else if cp.PriceUsd <= (1-conf.Config.Analyze.MaxRate)*sp.PriceUsd {
				smt = cp.MillUnixTime
				epUsd = (1 - conf.Config.Analyze.MaxRate) * sp.PriceUsd
				break
			}
		}
		fee = 20*conf.Config.Analyze.BuyFeeRate*contractUsd/sp.PriceUsd + 20*conf.Config.Analyze.SaleFeeRate*contractUsd/epUsd
		profit := 20*(contractUsd/sp.PriceUsd-contractUsd/epUsd) - fee
		endBalance := lastBalance + buyAmount + profit

		nco := &models.ContractOrder{
			Date:         date,
			Symbol:       conf.Config.Analyze.Symbol,
			BatchID:      batchID,
			Depth:        depth,
			StartBalance: lastBalance,
			EndBalance:   endBalance,
			CoinAmount:   coinAmount,
			ContractNum:  contractNum,
			BuyPrice:     sp.PriceUsd,
			SalePrice:    epUsd,
			BuyUsd:       buyUsd,
			BuyMillTime:  st,
			SaleMillTime: smt,
			MaxRate:      conf.Config.Analyze.MaxRate,
			Rate:         (epUsd - sp.PriceUsd) / sp.PriceUsd,
			Fee:          fee,
			Profit:       profit,
		}
		err = nco.Save()
		if err != nil {
			logger.Error("save contract order err: ", err)
		}
	}
	printSum()
}

func printSum() {
	cos, err := models.GetAllContractOrders(conf.Config.Analyze.Symbol)
	if err != nil {
		logger.Error("get contracts err: ", err)
	}
	sumBuyUsd := 0.0
	endBalance := 0.0
	sumProfit := 0.0
	maxDepth := 1
	sumFee := 0.0
	maxCoinAmount := 0.0
	for k, v := range cos {
		sumBuyUsd += v.BuyUsd
		sumProfit += v.Profit
		if k == len(cos)-1 {
			endBalance = v.EndBalance
		}
		if v.Depth > maxDepth {
			maxDepth = v.Depth
		}
		sumFee += v.Fee
		if v.CoinAmount > maxCoinAmount {
			maxCoinAmount = v.CoinAmount
		}
	}
	logger.Info("sum buy usd: ", sumBuyUsd, ", end banlance: ", endBalance, ", sum profit: ", sumProfit, ", max depth: ", maxDepth, ", sum fee: ", sumFee, ", max coin amount: ", maxCoinAmount)
}
