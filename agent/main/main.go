package main

import (
	"os"
	"log"
	"net"
	"fmt"
	"flag"
	"easyconfig/common"
	"easyconfig/core"
	"easyconfig/protocol"
	"easyconfig/agent/logic"
	"github.com/larspensjo/config"
	"time"
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
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9901")
	if err != nil {
		log.Fatalf("can's resolve udp address 127.0.0.1:9901\n")
		os.Exit(1)
	}
	//创建TCP客户端去连接broker
	addrStr := fmt.Sprintf("%s:%d", logic.GConf.BrokerIp, logic.GConf.BrokerPort)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addrStr)
	log.Printf("connecting to broker %s\n", addrStr)
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalf("can's connect tcp address %s\n", addrStr)
		os.Exit(1)
	}
	log.Printf("connected to broker %s\n", addrStr)

	ch := make(chan protocol.PullConfigReq, 1000)
	//routine: 向broker拉取最新配置
	go logic.PullFromBroker(ch, tcpConn)
	//routine: 接收来自broker的最新配置
	go logic.ReceiveFromBroker(controller, tcpConn)
	//routine: 周期性更新bucket信息
	go logic.PeriodicPull(controller, ch)
	//主goroutine负责服务于客户端

	time.Sleep(time.Second * 3)
	tcpConn.Close()

	time.Sleep(time.Second * 3000)
	logic.ClientService(udpAddr, ch)
}