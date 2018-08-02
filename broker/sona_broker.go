package main

import (
    "fmt"
    "flag"
    "sona/common"
    "sona/broker/conf"
    "sona/broker/logic"
)

func main() {
    configPath := flag.String("c", "", "broker configure file path")
    flag.Parse()

    if *configPath == "" {
        fmt.Println("broker configure file path is not specified")
        return
    }

    common.PrintLogo()
    conf.LoadGlobalConfig(*configPath)
    //启动broker server服务于agent
    go logic.StartBrokerService()
    //启动另一个服务，用于服务于管理端事务操作
    go logic.StartAdminService()
    //主G负责周期性拉最新数据

    expiredTs := int64(conf.GlobalConf.CacheExpiredTime)
    logic.CacheLayer.ClearExpired(expiredTs)
}
