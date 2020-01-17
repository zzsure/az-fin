package main

import (
	"github.com/urfave/cli"
	"az-fin/cmd/server"
	"az-fin/cmd/tool"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "az-fin"
	app.Commands = []cli.Command{
		server.Server,
		tool.InitDB,
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
