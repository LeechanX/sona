package main

import (
	"sona/common"
	"sona/broker/logic"
	"time"
)

func main() {
    common.PrintLogo()
    logic.LoadSelfConfig()
    //启动broker server服务于agent
    go logic.BrokerService()
    //启动另一个服务，用于服务于管理端事务操作
    go logic.AdminService()
    //TODO 主G逻辑未决
    time.Sleep(1000*time.Second)
}