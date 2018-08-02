package main

import (
    "sona/common"
    "sona/broker/conf"
    "sona/broker/logic"
)

func main() {
    common.PrintLogo()
    conf.LoadGlobalConfig()
    //启动broker server服务于agent
    go logic.StartBrokerService()
    //启动另一个服务，用于服务于管理端事务操作
    go logic.StartAdminService()
    //主G负责周期性拉最新数据
    expiredTs := int64(conf.GlobalConf.CacheExpiredTime)
    logic.CacheLayer.ClearExpired(expiredTs)
}