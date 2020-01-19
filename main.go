package main

import (
	"az-fin/cmd/server"
	"az-fin/cmd/tool"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "az-fin"
	app.Commands = []cli.Command{
		server.Server,
		tool.InitDB,
		tool.History,
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
