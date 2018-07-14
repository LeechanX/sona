package main

import (
    "sona/common"
    "sona/broker/logic"
    "time"
    "fmt"
)

func main() {
    common.PrintLogo()
    logic.LoadSelfConfig()
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