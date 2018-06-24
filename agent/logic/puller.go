package logic

import (
	"log"
	"net"
	"easyconfig/protocol"
)

func pullConfig(tcpConn *net.TCPConn, req *protocol.PullConfigReq) error {
	data := protocol.EncodeMessage(protocol.MsgTypeId_PullConfigReqId, req)
	_, err := tcpConn.Write(data)
	return err
}

//接收channel消息，向broker拉取
func PullFromBroker(ch <-chan protocol.PullConfigReq, tcpConn *net.TCPConn) {
	select {
	case req, ok := <- ch:
		if !ok {
			log.Fatalln("channel is closed, so routine exits")
			return
		}
		if err := pullConfig(tcpConn, &req);err != nil {
			//TODO: tcp handle error
		}
	}
}