package main

import (
    "log"
    "time"
    "sona/core"
    "sona/common"
    "sona/agent/conf"
    "sona/agent/logic"
)

func main() {
    common.PrintLogo()
    conf.LoadGlobalConfig()
    //创建共享内存控制者
    controller, err := core.GetConfigController()
    if err != nil {
        log.Printf("create configure controller meet error: %s\n", err)
        return
    }
    logic.ConfController = controller
    defer logic.ConfController.Close()

    //创建UDP服务于biz
    bizServer, err := logic.CreateBizServer("127.0.0.1", conf.GlobalConf.AgentPort)
    if err != nil {
        log.Printf("create biz server meet error: %s\n", err)
        return
    }
    logic.BizServer = bizServer
    defer logic.BizServer.Close()

    //创建broker客户端
    logic.BrokerClient = logic.CreateBrokerClient(conf.GlobalConf.BrokerIp, conf.GlobalConf.BrokerPort, true)

    //创建时间戳管理
    allServiceKeysMap := controller.GetAllServiceKeys()
    allServiceKeys := make([]string, 0)
    for serviceKey := range allServiceKeysMap {
        allServiceKeys = append(allServiceKeys, serviceKey)
    }
    logic.AccessRecordTable = logic.GetAccessRecord(allServiceKeys)

    //启动biz服务
    logic.BizServer.Start()

    //协程：周期性更新配置
    go logic.PeriodPulling()

    //协程：周期性清理无用serviceKey
    go logic.AccessRecordTable.Cleaner(controller)

    for {
        if err = logic.BrokerClient.Connect();err != nil {
            //重连
            time.Sleep(10 * time.Second)
            continue
        }
        //立刻重拉，以便告知broker 自己订阅了哪些serviceKey
        go logic.PullWhenStart()

        //连接已建立，一直运行，等待断连，说明网络出了问题
        logic.BrokerClient.Wait()
        time.Sleep(1 * time.Second)
    }
}