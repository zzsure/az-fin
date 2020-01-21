package tool

import (
	"az-fin/conf"
	"az-fin/library/db"
	"az-fin/library/log"
	"az-fin/models"
	"github.com/urfave/cli"
)

var InitDB = cli.Command{
	Name:  "init",
	Usage: "az-fin init db",
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
	Action: runInitDB,
}

func runInitDB(c *cli.Context) {
	conf.Init(c.String("conf"), c.String("args"))
	log.Init()
	db.Init()
	db.DB.LogMode(conf.Config.Database.LogMode)

	// TODO: 改为传参
	if true {
		models.MigrateTable()
	} else {
		models.CreateTable()
	}
}
