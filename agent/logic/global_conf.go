package logic

import (
	"os"
	"log"
	"flag"
	"github.com/larspensjo/config"
)

type GlobalConf struct {
	BrokerIp string
	BrokerPort int
	AgentPort int
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
	GConf.BrokerIp, err = cfg.String("broker", "ip")
	if err != nil {
		log.Panicf("configure broker-ip format error: %s\n", err)
		os.Exit(1)
	}
	GConf.BrokerPort, err = cfg.Int("broker", "port")
	if err != nil {
		log.Panicf("configure broker-port format error: %s\n", err)
		os.Exit(1)
	}

	if !cfg.HasSection("agent") {
		log.Panicln("configure has no section: agent")
		os.Exit(1)
	}
	GConf.AgentPort, err = cfg.Int("agent", "port")
	if err != nil {
		log.Panicf("configure agent-port format error: %s\n", err)
		os.Exit(1)
	}
}