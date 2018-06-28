package logic

import (
	"os"
	"log"
	"net"
	"sync/atomic"
	"easyconfig/core"
	"easyconfig/protocol"
	"github.com/golang/protobuf/proto"
)

//哪个client想订阅哪个service
type SubscribeMsg struct {
	serviceKey string
	addr *net.UDPAddr
}

type SubscribeResult struct {
	serviceKey string
	code int32//订阅结果
}

type ClientService struct {
	status int32
	conn *net.UDPConn
	sendQueue chan interface{}
}

//创建一个面向client服务的UDP服务器
func CreateClientService(udpAddr *net.UDPAddr) *ClientService {
	service := ClientService{}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Panicf("error listening udp: %s\n", err)
		os.Exit(1)
	}
	service.conn = conn
	service.status = 1
	service.sendQueue = make(chan interface{}, 100)
	return &service
}

func (cs *ClientService) Close() {
	if !atomic.CompareAndSwapInt32(&cs.status, 1, 0) {
		//已关闭
		return
	}
	cs.sendQueue<- nil
	cs.conn.Close()
}

func (cs *ClientService) Receiver(controller *core.ConfigController, c *Connection) {
	defer cs.Close()
	for {
		cmdId, cliAddr, pbData, err := protocol.DecodeUDPMessage(cs.conn)
		if err != nil {
			log.Panicf("%s\n", err)
			continue
		}

		if cmdId == protocol.MsgTypeId_SubscribeReqId {
			//client向agent订阅配置
			req := protocol.SubscribeReq{}
			if err = proto.Unmarshal(pbData, &req);err != nil {
				log.Printf("receive from udp data format error: %s\n", err)
				continue
			}

			//告知sender有新的订阅关系
			cs.sendQueue<- &SubscribeMsg{
				serviceKey:*req.ServiceKey,
				addr: cliAddr,
			}

			if controller.ExistService(*req.ServiceKey) {
				//已经有了，直接回复订阅成功
				cs.sendQueue<- &SubscribeResult{
					serviceKey:*req.ServiceKey,
					code: 0,
				}
			} else {
				//发送给拉取goroutine
				if atomic.LoadInt32(&c.status) == kConnStatusDisconnected {
					//确认与broker建立连接才发送
					log.Println("Client Service Routine: try to send request to puller goroutine")
					c.sendQueue<- &req
					log.Println("Client Service Routine: send request to puller goroutine ok")
				}
			}
		}
	}
}

func (cs *ClientService) Sender() {
	defer cs.Close()
	//保存订阅者与订阅key的关系
	relationship := make(map[string]map[*net.UDPAddr]bool)
	for {
		select {
		case msg, ok := <-cs.sendQueue:
			if !ok {
				//impossible code
				return
			}
			if msg == nil {
				return
			}

			if sub, ok := msg.(*SubscribeMsg);ok {
				//发来了订阅关系，保存一下
				if _, ok := relationship[sub.serviceKey];!ok {
					relationship[sub.serviceKey] = make(map[*net.UDPAddr]bool)
				}
				relationship[sub.serviceKey][sub.addr] = true
			} else if res, ok := msg.(*SubscribeResult);ok {
				//订阅的响应，则回复给每个订阅者
				if _, ok := relationship[res.serviceKey];ok {
					for addr := range relationship[res.serviceKey] {
						rsp := protocol.SubscribeAgentRsp{}
						rsp.ServiceKey = &res.serviceKey
						*rsp.Code = res.code
						data := protocol.EncodeMessage(protocol.MsgTypeId_SubscribeAgentRspId, &rsp)
						//TODO: 超时和错误处理
						cs.conn.WriteToUDP(data, addr)
					}
					//删除订阅关系
					delete(relationship, res.serviceKey)
				}
			}
		}
	}
}
