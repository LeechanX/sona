package logic

import (
	"easyconfig/protocol"
)

//接收channel消息，向broker拉取
func PullFromBroker(c *Connection) {
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
		data := protocol.EncodeMessage(protocol.MsgTypeId_PullConfigReqId, &req)
		if _, err := c.conn.Write(data);err != nil {
			return
		}
	}
}