package logic

import (
	"net"
	"log"
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
	go sender(&connection)
	//启动接收G
	go receiver(&connection)
}

//订阅
//true表示订阅成功，false表示之前已订阅
func (c *Connection) Subscript(configKey string) bool {
	if _, ok := c.subscriptList[configKey];ok {
		return false
	}
	c.subscriptList[configKey] = true
	return true
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
	for configKey := range c.subscriptList {
		Subscribed.UnSubscribed(configKey, c)
	}
	//为防止写channel产生panic，不关闭channel，仅发nil
	c.sendQueue<- nil
	c.conn.Close()
	atomic.AddInt32(&numberOfConnections, -1)
}

//推送被修改的配置
func (c *Connection) PushUpdatedData(key string, value string) {
	if !c.IsClosed() {
		c.sendQueue<- &protocol.PushConfigReq{
			Key:&key,
			Value:&value,
		}
	}
}

//推送被删除的配置
func (c *Connection) PushDeletedData(key string) {
	if !c.IsClosed() {
		c.sendQueue<- &protocol.RemoveConfigReq{
			Key:&key,
		}
	}
}

//接收消息的goroutine
func receiver(c *Connection) {
	//可能是网络出错，于是调用CloseConnect会主动关闭连接
	//也可能是其他G关闭了连接，这时调用Close将什么也不干
	defer c.Close()

	for {
		cmdId, pbData, err := protocol.DecodeTCPMessage(c.conn)
		if err != nil {
			log.Panicf("%s\n", err)
			return
		}
		if cmdId == protocol.MsgTypeId_PullConfigReqId {
			//agent向broker获取路由
			req := protocol.PullConfigReq{}
			err := proto.Unmarshal(pbData, &req)
			if err != nil {
				log.Panicf("receive from agent data format error: %s\n", err)
				return
			}
			//更新订阅列表、被订阅列表
			for _, key := range req.Keys {
				if c.Subscript(key) {
					//订阅成功
					Subscribed.Subscribed(key, c)
				}
			}

			rsp := protocol.PullConfigRsp{}
			var changed bool
			//从数据中查看配置信息
			for idx, key := range req.Keys {
				agentValue := req.Values[idx]
				localValue := ConfigData.GetData(key)
				if agentValue != localValue {
					changed = true
					//值有变化（含本地不存在的情况），回复
					rsp.Keys = append(rsp.Keys, key)
					rsp.Values = append(rsp.Values, localValue)
				}
			}
			if changed && !c.IsClosed() {
				//告知sender G
				c.sendQueue<- &rsp
			}
		} else {
			log.Printf("unknown request cmd id: %d\n", cmdId)
		}
	}
}

//发送消息的goroutine
func sender(c *Connection) {
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
			case protocol.PullConfigRsp:
				//回复包
				cmdId = protocol.MsgTypeId_PullConfigRspId
			case protocol.RemoveConfigReq:
				//发起删除命令
				cmdId = protocol.MsgTypeId_RemoveConfigReqId
			case protocol.PushConfigReq:
				//推送配置
				cmdId = protocol.MsgTypeId_PushConfigReqId
			default:
				return
			}

			data := protocol.EncodeMessage(cmdId, req)
			if _, err := c.conn.Write(data);err != nil {
				return
			}
		}
	}
}