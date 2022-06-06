package main

import (
	"context"
	"github.com/quanxiang-cloud/entrepot/api/restful"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"

	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/quanxiang-cloud/cabin/logger"
)

var (
	configPath = flag.String("config", "../configs/config.yml", "-config 配置文件地址")
)

func main() {
	flag.Parse()

	conf, err := config.NewConfig(*configPath)
	if err != nil {
		panic(err)
	}

	logger.Logger = logger.New(&conf.Log)
	ctx, cancel := context.WithCancel(context.Background())
	//start the router
	router, err := restful.NewRouter(ctx, conf)
	if err != nil {
		panic(err)
	}
	go router.Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			cancel()
			router.Close()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
