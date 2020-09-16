package tool

import (
	"az-fin/cmd"
	"az-fin/conf"
	"az-fin/consts"
	"az-fin/library/db"
	"az-fin/library/log"
	"az-fin/library/util"
	"az-fin/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/urfave/cli"
	"math"
	"math/rand"
	"strconv"
)

var f *excelize.File

// 1: 分析初始合约x张，最大接盘金额，盈利数据
// 2: 空头，不限制固定时间买入和固定时间卖出，随机买入后，波动r，平仓后随机买，设置最大深度
// 3：多头，策略与空头一样
// 4：随机买，每次只开20张合约，根据最大金额/10，得出买入次数，根据数据库条数除以次数得出随机范围
// 5：分析周一到周日哪天买平均价格值最小的次数最多
// 6：分析周日随机买入x$后上浮动f%卖出，手续费为r%会最大收益，本金和利润分别是多少
// 7：分析一段时间内，每日小时购买的最低数最大的值
// 8：每日凌晨购买，合适的卖出的浮动比例
// 9: 分析周一到周日小时时间买入利润最高，[0-5000):50$,[5000,10000):40$,[10000,15000):30$,[15000,20000):20$,[20000,*):10$

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

	// TODO: check config.toml params
	// 限制startTimes[0]最小，endTimes[len(endTimes)-1]最大

	analyzeType := c.Int("type")
	logger.Info("analyze type is: ", analyzeType)

	symbols := conf.Config.Analyze.Symbols
	startTimes := conf.Config.Analyze.StartTimes
	endTimes := conf.Config.Analyze.EndTimes

	f = excelize.NewFile()
	k := 0
	for i := 0; i < len(symbols); i++ {
		st, err := util.GetMillTimeByDate(startTimes[0])
		if err != nil {
			logger.Error("get mill date err: ", err)
			continue
		}

		et, err := util.GetMillTimeByDate(endTimes[len(endTimes)-1])
		if err != nil {
			logger.Error("get mill date err: ", err)
			continue
		}

		prices, err := models.GetPricesBySymbolAndTime(symbols[i], st, et)
		if err != nil {
			logger.Error("get symbol", symbols[i], " prices "+err.Error())
			continue
		}
		logger.Info("get symbol ", symbols[i], " prices count: ", len(prices))

		priceMap := make(map[int64]*models.Price, len(prices))
		for _, p := range prices {
			priceMap[p.MillUnixTime] = p
		}

		for j := 0; j < len(startTimes); j++ {
			startTime, err := util.GetMillTimeByDate(startTimes[j])
			if err != nil {
				logger.Error("get ", startTimes[j], " mill time err: "+err.Error())
				continue
			}

			endTime, err := util.GetMillTimeByDate(endTimes[j])
			if err != nil {
				logger.Error("get ", endTimes[j], " mill time err: "+err.Error())
				continue
			}

			var cos []*models.ContractOrder
			switch analyzeType {
			case consts.ANALYZE_BEAR_ORDER_FIX_BUY_HOUR:
				//fixBuyHour()
			case consts.ANALYZE_BEAR_ORDER_MAX_DEPTH, consts.ANALYZE_MORE_ORDER_MAX_DEPTH:
				for hour := 1; hour <= conf.Config.Analyze.MaxRandomHour; hour++ {
					k++
					cos = bearOrderMaxDepth(priceMap, startTime, endTime, hour, analyzeType)
					printProfitCosToExcel(k, hour, conf.Config.Analyze.MaxRate, symbols[i], startTimes[j], endTimes[j], cos)
				}
			case consts.ANALYZE_RANDOM_BUY:
				randomBuy()
			case consts.ANALYZE_WEEKLY_LOW_PRICE:
				// 开始时间从一个周一开始
				weeklyLowPrice(symbols[i], priceMap, startTime, endTime)
			case consts.ANALYZE_SUNNDAY_RANDOM_BUY:
				db.DB.Delete(models.Order{})
				db.DB.Delete(models.Profit{})
				_ = sunndayRandomBuy(symbols[i], priceMap, startTime, endTime)
			case consts.ANALYZE_DAILY_HOUR:
				_ = dailyHour(symbols[i], priceMap, startTime, endTime)
			case consts.ANALYZE_APPRO_RATE:
				for rate := 0.01; rate <= conf.Config.Analyze.MaxRate; rate += 0.01 {
					k++
					cos = approRate(priceMap, startTime, endTime, rate, analyzeType)
					printProfitCosToExcel(k, 1, rate, symbols[i], startTimes[j], endTimes[j], cos)
				}
			case consts.ANALYZE_AUTO_INVEST:
				autoInvest(symbols[i], priceMap, startTime, endTime)
			}
		}
		//break
	}
	if err := f.SaveAs(consts.DATA_BASE_DIR + "data.xlsx"); err != nil {
		logger.Error("err: ", err)
	}
}

func autoInvest(symbol string, priceMap map[int64]*models.Price, startTime, endTime int64) {
	buyMap := make(map[string]bool)
	amountMap := make(map[string]float64)
	symbolMap := make(map[string]float64)
	for umt := startTime; umt < endTime; umt += 60 * 60 * 1000 {
		sp, ok := priceMap[umt]
		if !ok {
			continue
		}
		t := util.GetTimeByMillUnixTime(umt)
		day := util.GetDateByTime(t)
		key1 := fmt.Sprintf("%v-%v", t.Weekday(), t.Hour())
		key2 := fmt.Sprintf("%v-%v-%v", day, t.Weekday(), t.Hour())
		logger.Info("umt:", umt, ", key1: ", key1, ", key2:", key2)
		amount := 10.0
		if sp.PriceUsd < 5000 {
			amount = 50.0
		} else if sp.PriceUsd >= 5000 && sp.PriceUsd < 10000 {
			amount = 40.0
		} else if sp.PriceUsd >= 10000 && sp.PriceUsd < 15000 {
			amount = 30.0
		} else if sp.PriceUsd >= 15000 && sp.PriceUsd < 20000 {
			amount = 20.0
		} else {
			amount = 10.0
		}
		if _, ok = buyMap[key2]; !ok {
			amountMap[key1] += amount
			symbolMap[key1] += amount / sp.PriceUsd
			buyMap[key2] = true
		}
	}
	idx := 1
	for k, v := range amountMap {
		if _, ok := symbolMap[k]; ok {
			axis := "A" + strconv.Itoa(idx)
			f.SetSheetRow("Sheet1", axis, &[]interface{}{symbol, k, v, symbolMap[k]})
			idx++
			logger.Info(k, ",", v, ",", symbolMap[k])
		}
	}
}

func dailyHour(symbol string, priceMap map[int64]*models.Price, startTime, endTime int64) error {
	dayPriceMap := make(map[string]float64)
	dayHourMap := make(map[string]int)

	for umt := startTime; umt < endTime; umt += 60 * 60 * 1000 {
		sp, ok := priceMap[umt]
		if !ok {
			continue
		}

		t := util.GetTimeByMillUnixTime(umt)
		day := util.GetDateByTime(t)
		if _, ok := dayPriceMap[day]; ok {
			if sp.PriceUsd < dayPriceMap[day] {
				dayPriceMap[day] = sp.PriceUsd
				dayHourMap[day] = t.Hour()
			}
		} else {
			dayPriceMap[day] = sp.PriceUsd
			dayHourMap[day] = t.Hour()
		}
	}
	hourCntMap := make(map[int]int)
	for _, dayHour := range dayHourMap {
		hourCntMap[dayHour]++
	}
	for hour, cnt := range hourCntMap {
		logger.Info("hour: ", hour, "cnt: ", hour, cnt)
	}
	return nil
}

func sunndayRandomBuy(symbol string, priceMap map[int64]*models.Price, startTime, endTime int64) error {
	k := 0
	for rate := 0.11; rate <= 0.30; rate += 0.01 {
		perMoney := conf.Config.Analyze.BuyMoney
		var orders []*models.Order
		var profits []*models.Profit
		dayOrderMap := make(map[string]*models.Order)
		for umt := startTime; umt < endTime; umt += 60 * 1000 {
			sp, ok := priceMap[umt]
			if !ok {
				continue
			}
			t := util.GetTimeByMillUnixTime(umt)
			day := util.GetDateByTime(t)
			//if "Sunday" == t.Weekday().String() {
			if _, ok := dayOrderMap[day]; !ok {
				amount := conf.Config.Analyze.BuyMoney / sp.PriceUsd
				fee := sp.PriceUsd * amount * conf.Config.Analyze.BuyFeeRate
				amount -= amount * conf.Config.Analyze.BuyFeeRate
				order := &models.Order{
					StrategyID:    consts.ANALYZE_SUNNDAY_RANDOM_BUY,
					Money:         conf.Config.Analyze.BuyMoney,
					Price:         sp.PriceUsd,
					Amount:        amount,
					Fee:           fee,
					Type:          consts.OrderTypeBuy,
					Status:        consts.OrderStatusSuccess,
					Ts:            umt / 1000,
					ExternalID:    "",
					RefrencePrice: sp.PriceUsd,
				}
				//order.Save()
				orders = append(orders, order)
				dayOrderMap[day] = order
			}
			//}
			for _, o := range orders {
				if o.Type == consts.OrderTypeBuy && o.Status == consts.OrderStatusSuccess {
					if sp.PriceUsd/o.Price-1 > rate {
						o.Status = consts.OrderStatusSettle
						fee := sp.PriceUsd * o.Amount * conf.Config.Analyze.SaleFeeRate
						money := sp.PriceUsd*o.Amount - fee
						order := &models.Order{
							StrategyID:    consts.ANALYZE_SUNNDAY_RANDOM_BUY,
							Money:         money,
							Price:         sp.PriceUsd,
							Amount:        o.Amount,
							Fee:           fee,
							Type:          consts.OrderTypeSale,
							Status:        consts.OrderStatusSettle,
							Ts:            umt / 1000,
							ExternalID:    "",
							RefrencePrice: sp.PriceUsd,
						}
						//order.Save()
						orders = append(orders, order)
						p, err := settle(o, order)
						if err == nil {
							profits = append(profits, p)
						}
					}
				}
			}
		}
		sumProfit := 0.0
		unsaleMoney := 0.0
		unSaleBtc := 0.0
		for _, o := range orders {
			if o.Type == consts.OrderTypeBuy && o.Status == consts.OrderStatusSuccess {
				unSaleBtc += o.Amount
				unsaleMoney += o.Money
			}
		}
		for _, p := range profits {
			sumProfit += p.Profit
		}
		k++
		axis := "A" + strconv.Itoa(k)
		logger.Info("axis:", axis, ", symbol: ", symbol, ", sd: ", startTime, ", ed: ", endTime, ", per_money: ", perMoney, ", float_rate: ", rate, ", sum_profit: ", sumProfit, ", unsale_money: ", unsaleMoney, ", unsale_btc: ", unSaleBtc)
		f.SetSheetRow("Sheet1", axis, &[]interface{}{symbol, startTime, endTime, perMoney, rate, sumProfit, unsaleMoney, unSaleBtc})
	}

	return nil
}

// 结算买单和卖单
func settle(bo, so *models.Order) (*models.Profit, error) {
	bo.Status = consts.OrderStatusSettle
	//err := bo.Save()
	//if err != nil {
	//	return nil, err
	//}
	so.Status = consts.OrderStatusSettle
	//err = so.Save()
	//if err != nil {
	//	return nil, err
	//}
	fee := so.Fee + bo.Fee
	ids := []uint{bo.ID, so.ID}
	idsByte, _ := json.Marshal(ids)
	day := util.GetDateByTime(util.GetTimeByMillUnixTime(so.Ts))
	p := &models.Profit{
		StrategyID:  consts.ANALYZE_SUNNDAY_RANDOM_BUY,
		TotalAmount: conf.Config.Analyze.BuyMoney,
		Depth:       1,
		FloatRate:   conf.Config.Analyze.MaxRate,
		Capital:     bo.Money,
		InCome:      so.Money,
		Fee:         fee,
		Profit:      so.Money - bo.Money,
		Reason:      "sunnday_random_buy",
		Day:         day,
		Orders:      string(idsByte),
	}
	//err = p.Save()
	return p, nil
}

func weeklyLowPrice(symbol string, priceMap map[int64]*models.Price, startTime, endTime int64) {
	weekHourMap := make(map[string]int)
	for st := startTime; st < endTime; st += 7 * 24 * 60 * 60 * 1000 {
		dayPriceMap := make(map[string]float64)
		dayCountMap := make(map[string]int)
		for unixTime := st; unixTime < st+7*24*60*60*1000; unixTime += 60 * 1000 {
			sp, ok := priceMap[unixTime]
			if !ok {
				continue
			}
			t := util.GetTimeByMillUnixTime(sp.MillUnixTime)
			str := t.Weekday().String()
			dayPriceMap[str] += sp.PriceUsd
			dayCountMap[str] += 1
		}
		minDay := ""
		minPrice := math.MaxFloat64
		for k, v := range dayPriceMap {
			if _, ok := dayCountMap[k]; ok {
				avgPrice := v / float64(dayCountMap[k])
				if avgPrice < minPrice {
					minDay = k
					minPrice = avgPrice
				}
			}
		}
		if minDay != "" {
			weekHourMap[minDay]++
		}
	}
	idx := 0
	for k, v := range weekHourMap {
		idx++
		axis := "A" + strconv.Itoa(idx)
		f.SetSheetRow("Sheet1", axis, &[]interface{}{symbol, startTime, endTime, k, v})
	}
}

func randomBuy() {
	// 根据最大金额，
}

func approRate(priceMap map[int64]*models.Price, startTime, endTime int64, rate float64, contractType int) []*models.ContractOrder {
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
			if lastCo.Status == 1 && (sp.PriceUsd >= (1+rate)*lastCo.BuyPrice || sp.PriceUsd <= (1-rate)*lastCo.BuyPrice) {
				err := saleOrder(lastCo, sp, contractType)
				if err != nil {
					logger.Error("sale err: ", err)
				}
			} else if lastCo.Status == 2 {
				hour := lastCo.Depth * 4
				if (sp.MillUnixTime - int64(hour*60*60*1000)) > lastCo.SaleMillTime {
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
	}
	return cos
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
	if t == consts.ANALYZE_BEAR_ORDER_FIX_BUY_HOUR || t == consts.ANALYZE_BEAR_ORDER_MAX_DEPTH {
		// 看空
		co.Profit = 20*(contractUsd/price.PriceUsd-contractUsd/co.BuyPrice) - co.Fee
	} else if t == consts.ANALYZE_MORE_ORDER_MAX_DEPTH || t == consts.ANALYZE_APPRO_RATE {
		// 看多
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

//func fixBuyHour() {
//	prices, err := models.GetPricesBySymbolAndTime(conf.Config.Analyze.Symbol, conf.Config.Analyze.StartMillTime, conf.Config.Analyze.EndMillTime)
//	if err != nil {
//		logger.Error("get prices " + err.Error())
//		return
//	}
//	logger.Info("get prices len: ", len(prices))
//	priceMap := make(map[int64]*models.Price, len(prices))
//	for _, p := range prices {
//		priceMap[p.MillUnixTime] = p
//	}
//
//	t := util.GetTimeByMillUnixTime(conf.Config.Analyze.StartMillTime)
//	m := util.GetMorningUnixTime(t)
//	st := (m + int64(conf.Config.Analyze.BuyHour)*60*60) * 1000
//
//	for ; st < conf.Config.Analyze.EndMillTime; st += 24 * 60 * 60 * 1000 {
//		t = util.GetTimeByMillUnixTime(st)
//		date := util.GetDateByTime(t)
//		logger.Info("deal date: ", date)
//
//		sp, ok := priceMap[st]
//		if !ok {
//			logger.Error("not have price: ", st)
//			continue
//		}
//		et := st + int64(conf.Config.Analyze.MaxSaleHour-conf.Config.Analyze.BuyHour)*60*60*1000
//		ep, ok := priceMap[et]
//		if !ok {
//			logger.Error("not have price: ", et)
//			continue
//		}
//		epUsd := ep.PriceUsd
//		smt := et
//
//		contractNum := conf.Config.Analyze.InitContractNum
//		lastBalance := 0.0
//		buyAmount := 0.0
//		buyUsd := 0.0
//		batchID := 1
//		depth := 1
//		fee := 0.0
//
//		co, err := models.GetLastContractOrder(conf.Config.Analyze.Symbol)
//		if err == nil {
//			// 同batch_id交易亏损
//			//batchProfit := 0.0
//			//cos, err := models.GetContractOrdersByBatchID(conf.Config.Analyze.Symbol, co.BatchID)
//			//if err == nil {
//			//	for _, bco := range cos {
//			//		batchProfit += bco.Profit
//			//	}
//			//}
//			if co.Profit < 0.0 {
//				contractNum = 2 * co.ContractNum
//				batchID = co.BatchID
//				depth = co.Depth + 1
//			} else {
//				batchID = co.BatchID + 1
//			}
//			lastBalance = co.EndBalance
//		}
//
//		contractUsd := 10.0 * float64(contractNum/20)
//		coinAmount := contractUsd / sp.PriceUsd
//		if lastBalance*sp.PriceUsd < contractUsd {
//			buyUsd = contractUsd - lastBalance*sp.PriceUsd
//			buyAmount = buyUsd / sp.PriceUsd
//		}
//
//		for it := st; it <= et; it += 60 * 1000 {
//			cp, ok := priceMap[st]
//			if !ok {
//				continue
//			}
//			// 上涨止损
//			if cp.PriceUsd >= (1+conf.Config.Analyze.MaxRate)*sp.PriceUsd {
//				smt = cp.MillUnixTime
//				epUsd = (1 + conf.Config.Analyze.MaxRate) * sp.PriceUsd
//				break
//			} else if cp.PriceUsd <= (1-conf.Config.Analyze.MaxRate)*sp.PriceUsd {
//				smt = cp.MillUnixTime
//				epUsd = (1 - conf.Config.Analyze.MaxRate) * sp.PriceUsd
//				break
//			}
//		}
//		fee = 20*conf.Config.Analyze.BuyFeeRate*contractUsd/sp.PriceUsd + 20*conf.Config.Analyze.SaleFeeRate*contractUsd/epUsd
//		profit := 20*(contractUsd/sp.PriceUsd-contractUsd/epUsd) - fee
//		endBalance := lastBalance + buyAmount + profit
//
//		nco := &models.ContractOrder{
//			Date:         date,
//			Symbol:       conf.Config.Analyze.Symbol,
//			BatchID:      batchID,
//			Depth:        depth,
//			StartBalance: lastBalance,
//			EndBalance:   endBalance,
//			CoinAmount:   coinAmount,
//			ContractNum:  contractNum,
//			BuyPrice:     sp.PriceUsd,
//			SalePrice:    epUsd,
//			BuyUsd:       buyUsd,
//			BuyMillTime:  st,
//			SaleMillTime: smt,
//			MaxRate:      conf.Config.Analyze.MaxRate,
//			Rate:         (epUsd - sp.PriceUsd) / sp.PriceUsd,
//			Fee:          fee,
//			Profit:       profit,
//		}
//		err = nco.Save()
//		if err != nil {
//			logger.Error("save contract order err: ", err)
//		}
//	}
//	printSum()
//}
//
//func printSum() {
//	cos, err := models.GetAllContractOrders(conf.Config.Analyze.Symbol)
//	if err != nil {
//		logger.Error("get contracts err: ", err)
//	}
//	sumBuyUsd := 0.0
//	endBalance := 0.0
//	sumProfit := 0.0
//	maxDepth := 1
//	sumFee := 0.0
//	maxCoinAmount := 0.0
//	for k, v := range cos {
//		sumBuyUsd += v.BuyUsd
//		sumProfit += v.Profit
//		if k == len(cos)-1 {
//			endBalance = v.EndBalance
//		}
//		if v.Depth > maxDepth {
//			maxDepth = v.Depth
//		}
//		sumFee += v.Fee
//		if v.CoinAmount > maxCoinAmount {
//			maxCoinAmount = v.CoinAmount
//		}
//	}
//	logger.Info("sum buy usd: ", sumBuyUsd, ", end banlance: ", endBalance, ", sum profit: ", sumProfit, ", max depth: ", maxDepth, ", sum fee: ", sumFee, ", max coin amount: ", maxCoinAmount)
//}

func printProfitCosToExcel(k, hour int, rate float64, symbol, sd, ed string, cos []*models.ContractOrder) {
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
	f.SetSheetRow("Sheet1", axis, &[]interface{}{symbol, hour, sd, ed, rate, sumUsd, endBalance, profit, maxDepth, sumFee})
}
