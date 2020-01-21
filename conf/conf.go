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
		LogMode bool `required:"true" flagUsage:"是否打印SQL日志"`
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
		Interval       string `required:"true" flagUsage:"获取的币种的间隔，m1, m5, m15, m30, h1, h2, h6, h12, d1"`
		StartMillTime  int64  `required:"true" flagUsage:"获取数据的开始时间"`
		EndMillTime    int64  `required:"true" flagUsage:"获取数据的结束时间"`
		MaxSleepSecond int    `required:"true" flagUsage:"最长访问http休息时间"`
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
