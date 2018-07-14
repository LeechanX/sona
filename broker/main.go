package main

import (
    "time"
    "fmt"
    "sona/common"
    "sona/broker/conf"
    "sona/broker/logic"
)

func main() {
    common.PrintLogo()
    conf.LoadGlobalConfig()
    //加载全部配置
    err := logic.ConfigData.Reset()
    if err != nil {
        fmt.Printf("load data from mongodb: %s\n", err)
        return
    }
    //启动broker server服务于agent
    go logic.BrokerService()
    //启动另一个服务，用于服务于管理端事务操作
    go logic.AdminService()
    //TODO 主G逻辑未决
    time.Sleep(1000*time.Second)
}