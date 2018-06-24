package logic

import (
	"log"
	"net"
	"easyconfig/protocol"
)

//接收channel消息，向broker拉取
func PullFromBroker(ch <-chan protocol.PullConfigReq, conn *net.TCPConn) {
	select {
	case req, ok := <- ch:
		if !ok {
			log.Fatalln("channel is closed, so routine exits")
			return
		}
		data := protocol.EncodeMessage(protocol.MsgTypeId_PullConfigReqId, &req)
		if _, err := conn.Write(data);err != nil {
			//TODO: 错误处理
		}
	}
}