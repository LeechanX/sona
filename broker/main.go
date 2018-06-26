package main

import (
	"os"
	"log"
	"flag"
	"easyconfig/common"
	"easyconfig/broker/logic"
	"github.com/larspensjo/config"
)

func loadSelfConfig() {
	cfgPath := flag.String("conf", "conf/easy-config-agent.ini", "configure file path")
	flag.Parse()
	//加载配置文件
	cfg, err := config.ReadDefault(*cfgPath)
	if err != nil {
		log.Panicf("load configure path error: %s\n", err)
		os.Exit(1)
	}

	if !cfg.HasSection("broker") {
		log.Panicln("configure has no section: broker")
		os.Exit(1)
	}
	logic.GConf.BrokerPort, err = cfg.Int("broker", "port")
	if err != nil {
		log.Panicf("configure broker-port format error: %s\n", err)
		os.Exit(1)
	}

	logic.GConf.ConnectionLimit, err = cfg.Int("broker", "connection-limit")
	if err != nil {
		log.Panicf("configure broker-ConnectionLimit format error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	common.PrintLogo()
	loadSelfConfig()
	//启动broker server服务于agent
	logic.BrokerService()
	//启动另一个服务，用于服务于管理端事务操作

}