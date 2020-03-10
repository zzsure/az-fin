package tool

import (
	"az-fin/cmd"
	"az-fin/conf"
	"az-fin/consts"
	"az-fin/library/db"
	"az-fin/library/log"
	"az-fin/library/util"
	"az-fin/models"
	"errors"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/urfave/cli"
	"math/rand"
	"strconv"
)

var f *excelize.File

// 1: 分析初始合约x张，最大接盘金额，盈利数据
// 2: 空头，不限制固定时间买入和固定时间卖出，随机买入后，波动r，平仓后随机买，设置最大深度
// 3：多头，策略与空头一样
// 4：随机买，每次只开20张合约，根据最大金额/10，得出买入次数，根据数据库条数除以次数得出随机范围

var Analyze = cli.Command{
	Name:  "analyze",
	Usage: "az-fin analyze data",
	Flags: []cli.Flag{
		cmd.StringFlag("conf, c", "config.toml", "toml配置文件"),
		cmd.StringFlag("args, a", "", "cmd line args"),
		cmd.IntFlag("type, t", 0, "api type"),
	},
	Action: runAnalyze,
}

func runAnalyze(c *cli.Context) {
	conf.Init(c.String("conf"), c.String("args"))
	log.Init()
	db.Init()
	db.DB.LogMode(conf.Config.Database.LogMode)

	//startTimeArr := [4]int64{"1483200000000", 1514736000000, 1546272000000, 1512057600000}
	//endTimeArr := [4]int64{1514736000000, 1546272000000, 1577808000000, 1577808000000}
	symbolArr := [9]string{"BTC", "ETH", "XRP", "BCH", "LTC", "EOS", "BSV", "ETC", "TRX"}
	startTimeArr := [4]string{"2017-01-01", "2018-01-01", "2019-01-01", "2017-12-01"}
	endTimeArr := [4]string{"2018-01-01", "2019-01-01", "2020-01-01", "2020-01-01"}

	if conf.Config.Analyze.EndMillTime <= conf.Config.Analyze.StartMillTime {
		logger.Error("end mill time should greater than start mill time")
		return
	}

	if conf.Config.Analyze.MaxSaleHour <= conf.Config.Analyze.BuyHour {
		logger.Error("sale hour should greater than buy hour")
		return
	}

	t := c.Int("type")
	logger.Info("type now is: ", t)

	f = excelize.NewFile()
	k := 0
	for i := 0; i < 9; i++ {
		st, err := util.GetMillTimeByDate(startTimeArr[0])
		if err != nil {
			logger.Error("get mill date err: ", err)
			continue
		}
		et, err := util.GetMillTimeByDate(endTimeArr[3])
		if err != nil {
			logger.Error("get mill date err: ", err)
			continue
		}
		prices, err := models.GetPricesBySymbolAndTime(symbolArr[i], st, et)
		if err != nil {
			logger.Error("get prices " + err.Error())
			continue
		}
		logger.Info("get prices len: ", len(prices))
		priceMap := make(map[int64]*models.Price, len(prices))
		for _, p := range prices {
			priceMap[p.MillUnixTime] = p
		}
		for j := 0; j < 4; j++ {
			startTime, err := util.GetMillTimeByDate(startTimeArr[j])
			if err != nil {
				logger.Error("get mill time err: " + err.Error())
				continue
			}
			endTime, err := util.GetMillTimeByDate(endTimeArr[j])
			if err != nil {
				logger.Error("get mill time err: " + err.Error())
				continue
			}
			var cos []*models.ContractOrder
			switch t {
			case consts.CONTRACT_BEAR_ORDER_FIX_BUY_HOUR:
				fixBuyHour()
			case consts.CONTRACT_BEAR_ORDER_MAX_DEPTH, consts.CONTRACT_MORE_ORDER_MAX_DEPTH:
				for hour := 1; hour <= conf.Config.Analyze.MaxRandomHour; hour++ {
					k++
					cos = bearOrderMaxDepth(priceMap, startTime, endTime, hour, t)
					printProfitCosToExcel(k, hour, symbolArr[i], startTimeArr[j], endTimeArr[j], cos)
				}
			case consts.CONTRACT_RANDOM_BUY:
				randomBuy()
			}
		}
		//break
	}
	if err := f.SaveAs(consts.DATA_BASE_DIR + "data.xlsx"); err != nil {
		logger.Error("err: ", err)
	}
}

func randomBuy() {
	// 根据最大金额，
}

func bearOrderMaxDepth(priceMap map[int64]*models.Price, startTime, endTime int64, hour, contractType int) []*models.ContractOrder {
	logger.Info("start_time: ", startTime, ", end_time:", endTime)
	var cos []*models.ContractOrder
	for st := startTime; st < endTime; st += 60 * 1000 {
		t := util.GetTimeByMillUnixTime(st)
		date := util.GetDateByTime(t)

		sp, ok := priceMap[st]
		if !ok {
			//logger.Error("not have price: ", st)
			continue
		}

		buyUsd := 0.0
		if len(cos) == 0 {
			buyUsd = 10.0 * float64(conf.Config.Analyze.InitContractNum/20)
			co := buyOrder(date, 1, 1, conf.Config.Analyze.InitContractNum, 0.0, buyUsd, sp)
			cos = append(cos, co)
		} else {
			lastCo := cos[len(cos)-1]
			// 如果处于买的状态，需要判断是否卖
			if lastCo.Status == 1 && (sp.PriceUsd >= (1+conf.Config.Analyze.MaxRate)*lastCo.BuyPrice || sp.PriceUsd <= (1-conf.Config.Analyze.MaxRate)*lastCo.BuyPrice) {
				err := saleOrder(lastCo, sp, contractType)
				if err != nil {
					logger.Error("sale err: ", err)
				}
			} else if lastCo.Status == 2 && (sp.MillUnixTime-int64(hour*60*60*1000)) > lastCo.SaleMillTime {
				contractNum := conf.Config.Analyze.InitContractNum
				batchID := lastCo.BatchID
				depth := 1
				if lastCo.Profit < 0 && lastCo.Depth < conf.Config.Analyze.MaxDepth {
					contractNum = 2 * lastCo.ContractNum
					depth = lastCo.Depth + 1
				} else {
					batchID++
				}
				contractUsd := 10.0 * float64(contractNum/20)
				lastBalance := lastCo.EndBalance
				if lastBalance*sp.PriceUsd < contractUsd {
					buyUsd = contractUsd - lastBalance*sp.PriceUsd
				}
				co := buyOrder(date, batchID, depth, contractNum, lastBalance, buyUsd, sp)
				cos = append(cos, co)
			}
		}
	}
	return cos
}

func saleOrder(co *models.ContractOrder, price *models.Price, t int) error {
	co.SalePrice = price.PriceUsd
	co.Status = 2
	co.SaleMillTime = price.MillUnixTime
	co.Rate = (price.PriceUsd - co.BuyPrice) / co.BuyPrice

	contractUsd := 10.0 * float64(co.ContractNum/20)
	co.Fee = 20*conf.Config.Analyze.BuyFeeRate*contractUsd/co.BuyPrice + 20*conf.Config.Analyze.SaleFeeRate*contractUsd/price.PriceUsd
	if t == consts.CONTRACT_BEAR_ORDER_FIX_BUY_HOUR || t == consts.CONTRACT_BEAR_ORDER_MAX_DEPTH {
		co.Profit = 20*(contractUsd/price.PriceUsd-contractUsd/co.BuyPrice) - co.Fee
	} else if t == consts.CONTRACT_MORE_ORDER_MAX_DEPTH {
		co.Profit = 20*(contractUsd/co.BuyPrice-contractUsd/price.PriceUsd) - co.Fee
	} else {
		return errors.New("type is wrong")
	}

	buyAmount := co.BuyUsd / co.BuyPrice
	co.EndBalance = co.StartBalance + buyAmount + co.Profit
	co.RandomHour = 1 + rand.Intn(conf.Config.Analyze.MaxRandomHour)
	//_ = co.Save()
	return nil
}

func buyOrder(date string, batchID, depth, contractNum int, lastBalance, buyUsd float64, price *models.Price) *models.ContractOrder {
	contractUsd := 10.0 * float64(contractNum/20)
	coinAmount := contractUsd / price.PriceUsd
	nco := &models.ContractOrder{
		Date:         date,
		Symbol:       price.Symbol,
		BatchID:      batchID,
		Depth:        depth,
		StartBalance: lastBalance,
		EndBalance:   0.0,
		CoinAmount:   coinAmount,
		ContractNum:  contractNum,
		BuyPrice:     price.PriceUsd,
		SalePrice:    0.0,
		BuyUsd:       buyUsd,
		BuyMillTime:  price.MillUnixTime,
		SaleMillTime: 0.0,
		MaxRate:      conf.Config.Analyze.MaxRate,
		Rate:         0.0,
		Fee:          0.0,
		Profit:       0.0,
		Status:       1,
	}
	return nco
}

func fixBuyHour() {
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
			// 同batch_id交易亏损
			//batchProfit := 0.0
			//cos, err := models.GetContractOrdersByBatchID(conf.Config.Analyze.Symbol, co.BatchID)
			//if err == nil {
			//	for _, bco := range cos {
			//		batchProfit += bco.Profit
			//	}
			//}
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

func printProfitCosToExcel(k, hour int, symbol, sd, ed string, cos []*models.ContractOrder) {
	sumUsd := 0.0
	endBalance := 0.0
	profit := 0.0
	maxDepth := 1
	sumFee := 0.0
	maxContractAmount := 0.0
	for _, co := range cos {
		if co.Status == 2 {
			sumUsd += co.BuyUsd
			endBalance = co.EndBalance
			profit += co.Profit
			if co.Depth > maxDepth {
				maxDepth = co.Depth
			}
			sumFee += co.Fee
			if co.CoinAmount > maxContractAmount {
				maxContractAmount = co.CoinAmount
			}
		}
	}
	logger.Info("symbol: ", symbol, ", sd: ", sd, ", ed: ", ed, ", sum_usd: ", sumUsd, ", end_balance: ", endBalance, ", profit: ", profit, ", max_depth: ", maxDepth, ", sum_fee: ", sumFee)
	axis := "A" + strconv.Itoa(k)
	f.SetSheetRow("Sheet1", axis, &[]interface{}{symbol, sd, ed, hour, sumUsd, endBalance, profit, maxDepth, sumFee})
}
