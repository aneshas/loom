package main

import (
	"log"

	"github.com/aneshas/helloapp/config"
	"github.com/aneshas/helloapp/web"
	"github.com/aneshas/helloapp/web/controller"
	"github.com/aneshas/loom"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg := loom.MustLoadConfig[config.Config]("./config")

	conn, err := cfg.DBConn()
	check(err)

	deps := loom.NewDeps()

	loom.Add(deps, conn)

	l := loom.New(deps)

	controller.Register(l)

	web.ConfigureServer(l)
	web.ConfigureRoutes(l)

	check(l.Start(cfg.Host))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
