package main

import (
    "os"
    "log"
    "net"
    "fmt"
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
        log.Fatalln(err)
        os.Exit(1)
    }
    defer controller.Close()

    //创建UDP服务于client
    udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", conf.GlobalConf.AgentPort))
    if err != nil {
        log.Fatalf("can's resolve udp address 127.0.0.1:9901\n")
        os.Exit(1)
    }

    //创建TCP客户端去连接broker
    addrStr := fmt.Sprintf("%s:%d", conf.GlobalConf.BrokerIp, conf.GlobalConf.BrokerPort)
    brokerConnector := logic.CreateConnect()

    //创建时间戳管理
    allServiceKeysMap := controller.GetAllServiceKeys()
    allServiceKeys := make([]string, 0)
    for serviceKey := range allServiceKeysMap {
        allServiceKeys = append(allServiceKeys, serviceKey)
    }
    record := logic.GetAccessRecord(allServiceKeys)

    //创建UDP服务器
    clientService := logic.CreateClientService(udpAddr)
    //协程1：对客户端提供读取服务
    go clientService.Receiver(record, controller, brokerConnector)
    //协程2：对客户端提供回复服务
    go clientService.Sender()

    //协程3：周期性更新配置
    go logic.PeriodicPull(controller, brokerConnector)

    //协程4：周期性清理无用serviceKey
    go record.Cleaner(controller)

    for {
        if err = brokerConnector.ConnectToBroker(addrStr);err != nil {
            time.Sleep(10 * time.Second)
            continue
        }
        //连接已建立
        //立刻重拉，以便告知broker 自己订阅了哪些serviceKey
        go logic.PullWhenStart(controller, brokerConnector)
        //创建2个协程
        //协程5：向broker拉取配置
        brokerConnector.Wg.Add(1)
        go brokerConnector.Pulling()

        //协程6：接收来自broker的消息
        brokerConnector.Wg.Add(1)
        go brokerConnector.Receiving(controller, clientService)

        //等待到2个协程终止，说明网络出了问题
        brokerConnector.Wg.Wait()
    }
}