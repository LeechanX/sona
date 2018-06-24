package logic

import (
	"os"
	"log"
	"net"
	"easyconfig/protocol"
	"github.com/golang/protobuf/proto"
)

func ClientService(udpAddr *net.UDPAddr, ch chan<- protocol.PullConfigReq) {
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Panicf("error listening udp: %s\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	for {
		data := make([]byte, 1024)
		nBytes, _, err := conn.ReadFromUDP(data)
		if err != nil {
			log.Panicf("receive from udp: %s\n", err)
			continue
		}

		if nBytes <= 8 {
			log.Panicf("receive from udp length error: %d\n", nBytes)
			continue
		}
		cmdId, pbData, err := protocol.DecodeMessage(data[:nBytes])
		if err != nil {
			continue
		}

		if cmdId == protocol.MsgTypeId_GetConfigReqId {
			//client向agent获取配置
			getConfigReq := protocol.GetConfigReq{}
			err = proto.Unmarshal(pbData, &getConfigReq)
			if err != nil {
				log.Panicf("receive from udp data format error: %s\n", err)
				continue
			}
			//获取信息
			key := getConfigReq.GetKey()
			log.Printf("Client Service Routine: client want to get configure: %s\n", key)
			pullConfigReq := protocol.PullConfigReq{}
			pullConfigReq.Keys = []string{key}
			//发送给拉取goroutine
			log.Println("Client Service Routine: try to send request to puller goroutine")
			ch<- pullConfigReq
			log.Println("Client Service Routine: send request to puller goroutine ok")
		}
	}
}
