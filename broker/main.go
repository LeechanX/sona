package main

import (
	"easyconfig/common"
	"easyconfig/broker/logic"
)

func main() {
	common.PrintLogo()
	logic.LoadSelfConfig()
	//启动broker server服务于agent
	logic.BrokerService()
	//启动另一个服务，用于服务于管理端事务操作

}