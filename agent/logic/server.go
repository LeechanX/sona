package logic

import (
	"os"
	"log"
	"net"
	"easyconfig/protocol"
	"github.com/golang/protobuf/proto"
	"sync/atomic"
)

func ClientService(udpAddr *net.UDPAddr, c *Connection) {
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Panicf("error listening udp: %s\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	for {
		cmdId, pbData, err := protocol.DecodeUDPMessage(conn)
		if err != nil {
			log.Panicf("%s\n", err)
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
			if atomic.LoadUint32(&c.status) == CONNECTED {
				//确认与broker建立连接才发送
				log.Println("Client Service Routine: try to send request to puller goroutine")
				c.sendQueue<- pullConfigReq
				log.Println("Client Service Routine: send request to puller goroutine ok")
			}
		}
	}
}
