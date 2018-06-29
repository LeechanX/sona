package logic

import (
	"os"
	"log"
	"flag"
	"github.com/larspensjo/config"
)

type GlobalConf struct {
	BrokerPort int
	AdminPort int
	//agent连接个数限制
	AgentConnectionLimit int
}

var GConf GlobalConf

func LoadSelfConfig() {
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
	GConf.BrokerPort, err = cfg.Int("broker", "port")
	if err != nil {
		log.Panicf("configure broker-port format error: %s\n", err)
		os.Exit(1)
	}

	GConf.AgentConnectionLimit, err = cfg.Int("broker", "connection-limit")
	if err != nil {
		log.Panicf("configure broker-ConnectionLimit format error: %s\n", err)
		os.Exit(1)
	}

	if !cfg.HasSection("admin") {
		log.Panicln("configure has no section: broker")
		os.Exit(1)
	}
	GConf.AdminPort, err = cfg.Int("admin", "port")
	if err != nil {
		log.Panicf("configure admin-port format error: %s\n", err)
		os.Exit(1)
	}
}