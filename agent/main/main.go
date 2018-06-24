package main

import (
	"os"
	"log"
	"net"
	"fmt"
	"flag"
	"time"
	"easyconfig/core"
	"easyconfig/common"
	"easyconfig/agent/logic"
	"github.com/larspensjo/config"
)

func loadSelfConfig() {
	cfgPath := flag.String("conf", "../conf/easy-config-agent.ini", "configure file path")
	flag.Parse()
	//加载配置文件
	cfg, err := config.ReadDefault(*cfgPath)
	if err != nil {
		log.Panicf("load configure path error: %s\n", err)
		os.Exit(1)
	}

	if !cfg.HasSection("broker") {
		log.Panicln("configure has no section: broker")
		os.Exit(1)
	}
	logic.GConf.BrokerIp, err = cfg.String("broker", "ip")
	if err != nil {
		log.Panicf("configure broker-ip format error: %s\n", err)
		os.Exit(1)
	}
	logic.GConf.BrokerPort, err = cfg.Int("broker", "port")
	if err != nil {
		log.Panicf("configure broker-port format error: %s\n", err)
		os.Exit(1)
	}

	if !cfg.HasSection("agent") {
		log.Panicln("configure has no section: agent")
		os.Exit(1)
	}
	logic.GConf.AgentPort, err = cfg.Int("agent", "port")
	if err != nil {
		log.Panicf("configure agent-port format error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	common.PrintLogo()
	loadSelfConfig()
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

	//协程1：对客户端提供服务
	go logic.ClientService(udpAddr, brokerConnector)

	//协程2：周期性更新配置
	go logic.PeriodicPull(controller, brokerConnector)

	for {
		if err = brokerConnector.ConnectToBroker(addrStr);err != nil {
			time.Sleep(10 * time.Second)
			continue
		}
		//连接已建立
		//创建2个协程
		//协程3：向broker拉取配置
		brokerConnector.Wg.Add(1)
		go logic.PullFromBroker(brokerConnector)

		//协程4：接收来自broker的消息
		brokerConnector.Wg.Add(1)
		go logic.ReceiveFromBroker(controller, brokerConnector)

		//等待到2个协程终止，说明网络出了问题
		brokerConnector.Wg.Wait()
	}
}