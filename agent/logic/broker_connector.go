package logic

import (
	"net"
	"log"
	"sync"
	"time"
	"errors"
	"sync/atomic"
	"easyconfig/core"
	"easyconfig/protocol"
	"github.com/golang/protobuf/proto"
)

const (
	kConnStatusConnected = iota
	kConnStatusDisconnected
)

type Connection struct {
	conn *net.TCPConn
	status int32
	Wg sync.WaitGroup
	sendQueue chan proto.Message
}

//创建一个connect结构体
func CreateConnect() *Connection {
	return &Connection{
		conn:nil,
		status:kConnStatusDisconnected,
		sendQueue:make(chan proto.Message, 1000),
	}
}

//执行连接
func (c *Connection) ConnectToBroker(addrStr string) error {
	if atomic.LoadInt32(&c.status) == kConnStatusConnected {
		return errors.New("already connected with broker")
	}
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addrStr)
	log.Printf("connecting to broker %s\n", addrStr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Printf("can's connect tcp address %s\n", addrStr)
		return err
	}
	log.Printf("connected to broker %s\n", addrStr)
	c.conn = conn
	//设置状态为已连接
	atomic.StoreInt32(&c.status, kConnStatusConnected)
	return nil
}

//关闭连接
func (c *Connection) CloseConnect() {
	if !atomic.CompareAndSwapInt32(&c.status, kConnStatusConnected, kConnStatusDisconnected) {
		log.Println("already closed the connection with broker")
		return
	}
	log.Println("now close the connection with broker")
	//防止panic，不关闭管道，而是发送消息nil告知已无消息
	c.sendQueue<- nil
	//关闭连接
	c.conn.Close()
}

//读取broker消息
func (c *Connection) Receiving(controller *core.ConfigController, clientService *ClientService) {
	defer func() {
		//可能是网络出错，于是调用CloseConnect会主动关闭连接
		//也可能是其他G关闭了连接，这时调用Close将什么也不干
		c.CloseConnect()
		c.Wg.Done()
	}()
	for {
		cmdId, pbData, err := protocol.DecodeTCPMessage(c.conn)
		if err != nil {
			log.Printf("%s\n", err)
			return
		}
		if cmdId == protocol.MsgTypeId_SubscribeBrokerRspId {
			//broker回复agent的订阅请求
			req := protocol.SubscribeBrokerRsp{}
			if *req.Code == 0 {
				//订阅成功，更新本地
				controller.UpdateService(*req.ServiceKey, req.Keys, req.Values)
			} else {
				//订阅失败，说明broker没有此配置
				controller.Remove(*req.ServiceKey)
			}
			rsp := protocol.SubscribeAgentRsp{}
			rsp.ServiceKey = req.ServiceKey
			rsp.Code = req.Code
			//可能需要回复给client
			clientService.sendQueue<- &SubscribeResult{
				serviceKey:*req.ServiceKey,
				code:*req.Code,
			}
		} else if cmdId == protocol.MsgTypeId_AddConfigReqId {
			//broker向agent发起添加配置命令
			req := protocol.AddConfigReq{}
			if err = proto.Unmarshal(pbData, &req);err != nil {
				log.Printf("receive from broker data format error: %s\n", err)
				return
			}
			//执行新增
			if err = controller.Set(*req.ServiceKey, *req.Key, *req.Value);err != nil {
				log.Printf("Set get error: %s\n", err)
			}
		} else if cmdId == protocol.MsgTypeId_DelConfigReqId {
			//broker向client发起删除一个配置项命令
			req := protocol.DelConfigReq{}
			if err = proto.Unmarshal(pbData, &req);err != nil {
				log.Printf("receive from broker data format error: %s\n", err)
				return
			}
			//执行删除
			controller.RemoveOne(*req.ServiceKey, *req.Key)
		} else if cmdId == protocol.MsgTypeId_UpdateConfigReqId {
			//broker向client发起更新一个配置项的命令
			req := protocol.UpdateConfigReq{}
			if err = proto.Unmarshal(pbData, &req);err != nil {
				log.Printf("receive from broker data format error: %s\n", err)
				return
			}
			//执行更新
			if err = controller.Set(*req.ServiceKey, *req.Key, *req.Value);err != nil {
				log.Printf("Set get error: %s\n", err)
			}
		} else if cmdId == protocol.MsgTypeId_PullServiceConfigRspId {
			//broker回复agent一个服务的最新配置
			req := protocol.PullServiceConfigRsp{}
			if err = proto.Unmarshal(pbData, &req);err != nil {
				log.Printf("receive from broker data format error: %s\n", err)
				return
			}
			//执行更新
			controller.UpdateService(*req.ServiceKey, req.Keys, req.Values)
		} else {
			log.Printf("unknown request cmd id: %d\n", cmdId)
		}
	}
}

//接收channel消息，向broker拉取
func (c *Connection) Pulling() {
	defer func() {
		//可能是网络出错，于是调用CloseConnect会主动关闭连接
		//也可能是其他G关闭了连接，这时调用CloseConnect将什么也不干
		c.CloseConnect()
		c.Wg.Done()
	}()
	select {
	case req, ok := <- c.sendQueue:
		if !ok {
			//impossible code
			return
		} else if req == nil {
			//说明连接已经关闭
			return
		}
		var cmdId protocol.MsgTypeId
		switch req.(type) {
		case *protocol.PullServiceConfigReq:
			//要向broker拉取
			cmdId = protocol.MsgTypeId_PullServiceConfigReqId
		case *protocol.SubscribeAgentRsp:
			//要向broker订阅
			cmdId = protocol.MsgTypeId_SubscribeBrokerRspId
		}
		data := protocol.EncodeMessage(cmdId, req)
		//设置100ms的超时
		c.conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
		if _, err := c.conn.Write(data);err != nil {
			return
		}
	}
}