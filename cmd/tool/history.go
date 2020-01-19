package tool

import (
	"az-fin/conf"
	"az-fin/consts"
	"az-fin/library/db"
	"az-fin/library/log"
	"az-fin/library/util"
	"az-fin/library/util/net/http"
	"az-fin/models"
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/op/go-logging"
	"github.com/urfave/cli"
	"strconv"
	"strings"
)

var logger = logging.MustGetLogger("cmd/tool")

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
	db.DB.LogMode(true)
	if conf.Config.History.EndMillTime <= conf.Config.History.StartMillTime {
		logger.Error("end mill time should greater than start mill time")
		return
	}
	for t := conf.Config.History.StartMillTime; t < conf.Config.History.EndMillTime; t += 24 * 60 * 60 * 1000 {
		time := util.GetTimeByMillUnixTime(t)
		logger.Info("deal data time: ", util.GetFormatTime(time))
		priceResults, err := getPrices(conf.Config.History.CoinCapID, conf.Config.History.Interval, t, t+24*60*60*1000)
		if err != nil {
			logger.Error("get price result err: ", err)
			break
		}
		err = dealPrices(priceResults)
		if err != nil {
			logger.Error("save price err: ", err)
			break
		}
		util.RandomSleep(conf.Config.History.MaxSleepSecond)
	}
}

type PriceResult struct {
	PriceUsd string `json:"priceUsd"`
	Time     int64  `json:"time"`
}
type PriceResults []PriceResult

func getPrices(coinCapID, interval string, start, end int64) (PriceResults, error) {
	url := fmt.Sprintf("%s/%s/%s/history?interval=%s&start=%d&end=%d", strings.TrimRight(consts.COINCAP_URL, "/"), strings.TrimLeft(consts.ASSETS_URI, "/"), coinCapID, interval, start, end)
	logger.Info("get url: ", url)
	b, err := http.Get(url, nil)
	logger.Info("get price: ", string(b))
	if err != nil {
		return nil, err
	}
	data, _, _, _ := jsonparser.Get(b, "data")
	var priceResults PriceResults
	if err := json.Unmarshal(data, &priceResults); err != nil {
		return nil, err
	}
	return priceResults, nil
}

func dealPrices(priceResults PriceResults) error {
	for _, priceResult := range priceResults {
		p, err := strconv.ParseFloat(priceResult.PriceUsd, 64)
		if err != nil {
			return err
		}
		price := &models.Price{
			CoinCapID:    conf.Config.History.CoinCapID,
			Interval:     conf.Config.History.Interval,
			PriceUsd:     p,
			MillUnixTime: priceResult.Time,
		}
		err = price.Save()
		if err != nil {
			return err
		}
	}
	return nil
}
