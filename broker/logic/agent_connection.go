package logic

import (
	"net"
	"log"
	"time"
	"sync/atomic"
	"easyconfig/protocol"
	"github.com/golang/protobuf/proto"
)

const (
	kConnStatusConnected = int32(1)
	kConnStatusDisconnected
)

type Connection struct {
	status int32
	conn *net.TCPConn
	//此连接订阅了哪些configure
	subscriptList map[string]bool
	sendQueue chan proto.Message
}

//创建
func CreateConnection(c *net.TCPConn) {
	connection := Connection{
		conn:c,
		sendQueue:make(chan proto.Message, 1000),
	}
	atomic.AddInt32(&numberOfConnections, 1)
	//启动发送G
	go connection.sender()
	//启动接收G
	go connection.receiver()
}

//订阅
func (c *Connection) Subscribe(configKey string) {
	c.subscriptList[configKey] = true
}

//连接是否存活
func (c *Connection) IsClosed() bool {
	return atomic.LoadInt32(&c.status) == kConnStatusDisconnected
}

//关闭连接
func (c *Connection) Close() {
	if !atomic.CompareAndSwapInt32(&c.status, kConnStatusConnected, kConnStatusDisconnected) {
		//已被关闭
		return
	}
	if c.conn == nil {
		log.Println("already closed the connection with agent")
		return
	}
	//需要在被订阅列表里删除每个关联
	for serviceKey := range c.subscriptList {
		SubscribedBook.UnSubscribed(serviceKey, c)
	}
	//为防止写channel产生panic，不关闭channel，仅发nil
	c.sendQueue<- nil
	c.conn.Close()
	atomic.AddInt32(&numberOfConnections, -1)
}

//推送新增的、被修改的配置
func (c *Connection) PushAddOrUpdated(serviceKey string, confKey string, value string) {
	if !c.IsClosed() {
		c.sendQueue<- &protocol.AddConfigReq{
			ServiceKey:&serviceKey,
			Key:&confKey,
			Value:&value,
		}
	}
}

//推送被删除的配置
func (c *Connection) PushDeleted(serviceKey string, confKey string) {
	if !c.IsClosed() {
		c.sendQueue<- &protocol.DelConfigReq{
			ServiceKey:&serviceKey,
			Key:&confKey,
		}
	}
}

//接收消息的goroutine
func (c *Connection) receiver() {
	//可能是网络出错，于是调用CloseConnect会主动关闭连接
	//也可能是其他G关闭了连接，这时调用Close将什么也不干
	defer c.Close()
	for {
		cmdId, pbData, err := protocol.DecodeTCPMessage(c.conn)
		if err != nil {
			log.Panicf("%s\n", err)
			return
		}
		if cmdId == protocol.MsgTypeId_SubscribeReqId {
			//agent来订阅配置
			req := protocol.SubscribeReq{}
			if err := proto.Unmarshal(pbData, &req);err != nil {
				log.Panicf("receive from agent data format error: %s\n", err)
				return
			}
			rsp := protocol.SubscribeBrokerRsp{}
			rsp.ServiceKey = req.ServiceKey
			//查看是否有此配置
			configs := ConfigData.IsServiceKeyExist(*req.ServiceKey)
			if configs == nil {
				//订阅失败
				*rsp.Code = -1
			} else {
				//记录：serviceKey被c连接所订阅
				SubscribedBook.Subscribed(*req.ServiceKey, c)
				//记录：c连接订阅了serviceKey
				c.Subscribe(*req.ServiceKey)
				//订阅成功
				*rsp.Code = 0
				//填充配置
				for key, value := range configs {
					rsp.Keys = append(rsp.Keys, key)
					rsp.Values = append(rsp.Values, value)
				}
			}
			if !c.IsClosed() {
				//告知sender G
				c.sendQueue<- &rsp
			}
		}
		if cmdId == protocol.MsgTypeId_PullServiceConfigReqId {
			//agent向broker获取路由
			req := protocol.PullServiceConfigReq{}
			err := proto.Unmarshal(pbData, &req)
			if err != nil {
				log.Panicf("receive from agent data format error: %s\n", err)
				return
			}
			rsp := protocol.PullServiceConfigRsp{}
			rsp.ServiceKey = req.ServiceKey
			//查看是否有此配置
			configs := ConfigData.IsServiceKeyExist(*req.ServiceKey)
			if configs != nil {
				//有可能是连接建立后立刻来拉取的，主要意图是：
				//1、告知broker：agent已订阅哪些service keys
				//2、更新一下最新配置

				//记录：serviceKey被c连接所订阅
				SubscribedBook.Subscribed(*req.ServiceKey, c)
				//记录：c连接订阅了serviceKey
				c.Subscribe(*req.ServiceKey)
				//填充配置
				for key, value := range configs {
					rsp.Keys = append(rsp.Keys, key)
					rsp.Values = append(rsp.Values, value)
				}
			}
			if !c.IsClosed() {
				//告知sender G
				c.sendQueue<- &rsp
			}
		} else {
			log.Printf("unknown request cmd id: %d\n", cmdId)
		}
	}
}

//发送消息的goroutine
func (c *Connection) sender() {
	//可能是网络出错，于是调用CloseConnect会主动关闭连接
	//也可能是其他G关闭了连接，这时调用CloseConnect将什么也不干
	defer c.Close()
	for {
		select {
		case req, ok := <- c.sendQueue:
			if !ok {
				//impossible code: 已经关闭了
				return
			} else if req == nil {
				//连接已经关闭
				return
			}
			var cmdId protocol.MsgTypeId
			switch req.(type) {
			case protocol.AddConfigReq:
				//推送新的、更新的配置
				cmdId = protocol.MsgTypeId_AddConfigReqId
			case protocol.DelConfigReq:
				//推送删除配置
				cmdId = protocol.MsgTypeId_DelConfigReqId
			case protocol.PullServiceConfigRsp:
				//回复拉取配置
				cmdId = protocol.MsgTypeId_PullServiceConfigRspId
			default:
				return
			}

			data := protocol.EncodeMessage(cmdId, req)
			c.conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
			if _, err := c.conn.Write(data);err != nil {
				return
			}
		}
	}
}