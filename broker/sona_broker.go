package main

import (
    "log"
    "sona/common"
    "sona/broker/conf"
    "sona/broker/logic"
)

func main() {
    common.PrintLogo()
    conf.LoadGlobalConfig()
    //加载全部配置
    err := logic.CacheData.Load()
    if err != nil {
        log.Printf("load data from mongodb: %s\n", err)
        return
    } else {
        log.Println("load data from mongodb ok")
    }
    //启动broker server服务于agent
    go logic.StartBrokerService()
    //启动另一个服务，用于服务于管理端事务操作
    go logic.StartAdminService()
    //主G负责周期性拉最新数据
    logic.PeriodReload()
}