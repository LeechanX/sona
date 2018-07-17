package conf

import (
    "os"
    "log"
    "flag"
    "github.com/larspensjo/config"
)

type GlobalConfigure struct {
    BrokerPort int
    AdminPort int
    //agent连接个数限制
    AgentConnectionLimit int
    DbHost string
    DbPort int
    DbName string
    DbCollectionName string
}

var GlobalConf GlobalConfigure

func LoadGlobalConfig() {
    cfgPath := flag.String("conf", "conf/sona-broker.ini", "configure file path")
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
    GlobalConf.BrokerPort, err = cfg.Int("broker", "port")
    if err != nil {
        log.Printf("configure broker-port format error: %s\n", err)
        os.Exit(1)
    }

    GlobalConf.AgentConnectionLimit, err = cfg.Int("broker", "connection-limit")
    if err != nil {
        log.Printf("configure broker-ConnectionLimit format error: %s\n", err)
        os.Exit(1)
    }

    if !cfg.HasSection("admin") {
        log.Panicln("configure has no section: broker")
        os.Exit(1)
    }
    GlobalConf.AdminPort, err = cfg.Int("admin", "port")
    if err != nil {
        log.Printf("configure admin-port format error: %s\n", err)
        os.Exit(1)
    }

    if !cfg.HasSection("db") {
        log.Panicln("configure has no section: broker")
        os.Exit(1)
    }

    GlobalConf.DbHost, err = cfg.String("db","host")
    if err != nil {
        log.Printf("configure db-host format error: %s\n", err)
        os.Exit(1)
    }

    GlobalConf.DbPort, err = cfg.Int("db", "port")
    if err != nil {
        log.Printf("configure db-port format error: %s\n", err)
        os.Exit(1)
    }

    GlobalConf.DbHost, err = cfg.String("db","host")
    if err != nil {
        log.Printf("configure db-host format error: %s\n", err)
        os.Exit(1)
    }

    GlobalConf.DbName, err = cfg.String("db","database")
    if err != nil {
        log.Printf("configure db-database format error: %s\n", err)
        os.Exit(1)
    }

    GlobalConf.DbCollectionName, err = cfg.String("db","collection")
    if err != nil {
        log.Printf("configure db-collection format error: %s\n", err)
        os.Exit(1)
    }
}
