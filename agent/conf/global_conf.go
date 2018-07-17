package conf

import (
    "os"
    "log"
    "flag"
    "github.com/larspensjo/config"
)

type GlobalConfigure struct {
    BrokerIp string
    BrokerPort int
    AgentPort int
}

var GlobalConf GlobalConfigure

func LoadGlobalConfig() {
    cfgPath := flag.String("conf", "conf/sona-agent.ini", "configure file path")
    flag.Parse()
    //加载配置文件
    cfg, err := config.ReadDefault(*cfgPath)
    if err != nil {
        log.Printf("load configure path error: %s\n", err)
        os.Exit(1)
    }

    if !cfg.HasSection("broker") {
        log.Panicln("configure has no section: broker")
        os.Exit(1)
    }
    GlobalConf.BrokerIp, err = cfg.String("broker", "ip")
    if err != nil {
        log.Printf("configure broker-ip format error: %s\n", err)
        os.Exit(1)
    }
    GlobalConf.BrokerPort, err = cfg.Int("broker", "port")
    if err != nil {
        log.Printf("configure broker-port format error: %s\n", err)
        os.Exit(1)
    }

    if !cfg.HasSection("agent") {
        log.Panicln("configure has no section: agent")
        os.Exit(1)
    }
    GlobalConf.AgentPort, err = cfg.Int("agent", "port")
    if err != nil {
        log.Printf("configure agent-port format error: %s\n", err)
        os.Exit(1)
    }
}