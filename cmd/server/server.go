package server

import (
	"az-fin/conf"
	"az-fin/controller/v1"
	"az-fin/library/db"
	"az-fin/library/log"
	"az-fin/library/redis"
	"az-fin/middleware"
	"az-fin/modules/cron"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli"
)

var Server = cli.Command{
	Name:  "server",
	Usage: "az-fin http server",
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
	Action: run,
}

func run(c *cli.Context) {
	conf.Init(c.String("conf"), c.String("args"))
	log.Init()
	db.Init()
	redis.Init()
	cron.Init()

	_ = GinEngine().Run(conf.Config.Server.Listen)
}

func GinEngine() *gin.Engine {
	var r *gin.Engine
	if conf.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
		r.Use(middleware.Recovery)
	} else {
		r = gin.Default()
	}
	r.Use(middleware.Access)
	r.Use(middleware.Auth)
	r.GET("/health")
	V1(r)

	return r
}

func V1(r *gin.Engine) {
	g := r.Group("/v1")
	{
		g.POST("/echo", v1.Echo)
		g.GET("/price/excel", v1.PriceExcel)
		g.GET("/price/test", v1.PriceTest)
		g.GET("/price/list", v1.PriceList)
	}
}
