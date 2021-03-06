package conf

import (
	"az-fin/library/util"
	"github.com/koding/multiconfig"
	"strings"
	"time"
)

type ConfigTOML struct {
	Server struct {
		Listen             string         `required:"true" flagUsage:"服务监听地址"`
		Env                string         `default:"Pro" flagUsage:"服务运行时环境"`
		MaxHttpRequestBody int64          `default:"4" flagUsage:"最大允许的http请求body，单位M"`
		TimeLocation       *time.Location `flagUsage:"用于time.ParseInLocation"`
	}

	Auth struct {
		Secret  string            `flagUsage:"跳过鉴权的hack"`
		Account map[string]string `flagUsage:"复杂验证，apiKey=>apiSecret"`
	}

	Database struct {
		HostPort     string `required:"true" flagUsage:"数据库连接，eg：tcp(127.0.0.1:3306)"`
		UserPassword string `required:"true" flagUsage:"数据库账号密码"`
		DB           string `required:"true" flagUsage:"数据库"`
		Conn         struct {
			MaxLifeTime int `default:"600" flagUsage:"连接最长存活时间，单位s"`
			MaxIdle     int `default:"10" flagUsage:"最多空闲连接数"`
			MaxOpen     int `default:"80" flagUsage:"最多打开连接数"`
		}
		LogMode bool `default:"false" flagUsage:"是否打印SQL日志"`
	}

	Redis struct {
		IsUse    bool   `default:"true" flagUsage:"是否开启缓存"`
		Addr     string `flagUsage:"Redis地址"`
		Password string `default:"" flagUsage:"Redis密码"`
		DB       int    `default:"0" flagUsage:"Redis数据库"`
		PoolSize int    `default:"100"  flagUsage:"Redis连接池大小"`
	}

	Log struct {
		Type  string `default:"json" flagUsage:"日志格式，json|raw"`
		Level int    `default:"5" flagUsage:"日志级别：0 CRITICAL, 1 ERROR, 2 WARNING, 3 NOTICE, 4 INFO, 5 DEBUG"`
	} `flagUsage:"服务日志配置"`

	History struct {
		CoinCapID      string `required:"true" flagUsage:"获取的币种"`
		Symbol         string `required:"true" flagUsage:"获取的币种符号"`
		Interval       string `required:"true" flagUsage:"获取的币种的间隔，m1, m5, m15, m30, h1, h2, h6, h12, d1"`
		StartMillTime  int64  `required:"true" flagUsage:"获取数据的开始时间，毫秒"`
		EndMillTime    int64  `required:"true" flagUsage:"获取数据的结束时间，毫秒"`
		MaxSleepSecond int    `required:"true" flagUsage:"最长访问http休息时间"`
	}

	Contract struct {
		CoinCapID     string  `required:"true" flagUsage:"分析的币种"`
		StartMillTime int64   `required:"true" flagUsage:"分析数据的开始时间，毫秒"`
		EndMillTime   int64   `required:"true" flagUsage:"分析数据的结束时间，毫秒"`
		BuyHour       int     `required:"true" flagUsage:"当天买入时间，1-23"`
		MaxSaleHour   int     `required:"true" flagUsage:"当天最晚卖出时间，2-24"`
		MaxRate       float64 `required:"true" flagUsage:"最大幅度，0.02"`
	}

	Analyze struct {
		Symbols         []string `required:"true" flagUsage:"分析的币种，如BTC"`
		StartTimes      []string `required:"true" flagUsage:"分析数据的开始时间，毫秒"`
		EndTimes        []string `required:"true" flagUsage:"分析数据的结束时间，毫秒"`
		BuyMoney        float64  `required:"true" flagUsage:"每次买入金额"`
		BuyHour         int      `required:"true" flagUsage:"当天买入时间，1-23"`
		MaxSaleHour     int      `required:"true" flagUsage:"当天最晚卖出时间，2-24"`
		MaxRate         float64  `required:"true" flagUsage:"最大幅度，0.02"`
		BuyFeeRate      float64  `required:"true" flagUsage:"买入的费率"`
		SaleFeeRate     float64  `required:"true" flagUsage:"卖出的费率"`
		InitContractNum int      `required:"true" flagUsage:"初始合约张数"`
		MinRandomHour   int      `required:"true" flagUsage:"再次下单最小间隔小时"`
		MaxRandomHour   int      `required:"true" flagUsage:"卖出后最大等待小时"`
		MaxDepth        int      `required:"true" flagUsage:"最大深度"`
	}
}

func (c *ConfigTOML) IsProduction() bool {
	return strings.ToLower(c.Server.Env) == "pro"
}

var Config *ConfigTOML

func Init(tomlPath, args string) {
	var err error
	var loaders = []multiconfig.Loader{
		&multiconfig.TagLoader{},
		&multiconfig.TOMLLoader{Path: tomlPath},
		&multiconfig.EnvironmentLoader{},
	}
	m := multiconfig.DefaultLoader{
		Loader:    multiconfig.MultiLoader(loaders...),
		Validator: multiconfig.MultiValidator(&multiconfig.RequiredValidator{}),
	}
	Config = new(ConfigTOML)
	m.MustLoad(&Config)

	Config.Server.TimeLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}

	_ = util.PrettyPrint(Config)
}
