package main

import (
	"os"
	"log"
	"net"
	"fmt"
	"time"
	"easyconfig/core"
	"easyconfig/common"
	"easyconfig/agent/logic"
)

func main() {
	common.PrintLogo()
	logic.LoadSelfConfig()
	controller, err := core.GetConfigController()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	defer controller.Close()

	//创建UDP服务于client
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", logic.GConf.AgentPort))
	if err != nil {
		log.Fatalf("can's resolve udp address 127.0.0.1:9901\n")
		os.Exit(1)
	}

	//创建TCP客户端去连接broker
	addrStr := fmt.Sprintf("%s:%d", logic.GConf.BrokerIp, logic.GConf.BrokerPort)
	brokerConnector := logic.CreateConnect()

	//创建UDP服务器
	clientService := logic.CreateClientService(udpAddr)
	//协程1：对客户端提供读取服务
	go clientService.Receiver(controller, brokerConnector)
	//协程2：对客户端提供回复服务
	go clientService.Sender()

	//协程3：周期性更新配置
	go logic.PeriodicPull(controller, brokerConnector)

	for {
		if err = brokerConnector.ConnectToBroker(addrStr);err != nil {
			time.Sleep(10 * time.Second)
			continue
		}
		//连接已建立
		//立刻重拉，以便告知broker 自己订阅了哪些serviceKey
		go logic.PullWhenStart(controller, brokerConnector)
		//创建2个协程
		//协程4：向broker拉取配置
		brokerConnector.Wg.Add(1)
		go brokerConnector.Pulling()

		//协程5：接收来自broker的消息
		brokerConnector.Wg.Add(1)
		go brokerConnector.Receiving(controller, clientService)

		//等待到2个协程终止，说明网络出了问题
		brokerConnector.Wg.Wait()
	}
}