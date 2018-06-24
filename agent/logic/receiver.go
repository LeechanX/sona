package logic

import (
	"log"
	"net"
	"easyconfig/core"
	"easyconfig/protocol"
	"github.com/golang/protobuf/proto"
)

func ReceiveFromBroker(controller *core.ConfigController, tcpConn *net.TCPConn) {
	for {
		data := make([]byte, 1024)
		nBytes, err := tcpConn.Read(data)
		if err != nil {
			//TODO
			log.Panicf("tcp receive error: %s\n", err)
		}
		if nBytes <= 8 {
			log.Panicf("receive from broker length error: %d\n", nBytes)
			continue
		}
		cmdId, pbData, err := protocol.DecodeMessage(data[:nBytes])
		if err != nil {
			//TODO
		}
		//收到来自broker的回复
		if cmdId == protocol.MsgTypeId_PullConfigRspId {
			pullConfigRsp := protocol.PullConfigRsp{}
			err = proto.Unmarshal(pbData, &pullConfigRsp)
			if err != nil {
				log.Panicf("receive from broker data format error: %s\n", err)
				continue
			}

			if len(pullConfigRsp.Keys) != len(pullConfigRsp.Values) {
				continue
			}
			//依次更新每个结果
			for idx := range pullConfigRsp.Keys {
				key := pullConfigRsp.Keys[idx]
				log.Printf("get updated configure %s\n", key)
				value := pullConfigRsp.Values[idx]
				err = controller.Set(key, value)
				if err != nil {
					log.Panicf("receive from broker data format error: %s\n", err)
					continue
				}
			}
		} else if cmdId == protocol.MsgTypeId_PushConfigReqId {
			//broker主动推送配置到来
			pushConfigRsp := protocol.PushConfigReq{}
			err = proto.Unmarshal(pbData, &pushConfigRsp)
			if err != nil {
				log.Panicf("receive from broker data format error: %s\n", err)
				continue
			}
			//更新配置
			key := *pushConfigRsp.Key
			log.Printf("Receiver Routine: get updated configure %s\n", key)
			value := *pushConfigRsp.Value
			err = controller.Set(key, value)
			if err != nil {
				log.Panicf("Receiver Routine: receive from broker data format error: %s\n", err)
			}
		} else if cmdId == protocol.MsgTypeId_RemoveConfigReqId {
			//broker要求删除配置
			removeConfigReq := protocol.RemoveConfigReq{}
			err = proto.Unmarshal(pbData, &removeConfigReq)
			if err != nil {
				log.Panicf("receive from broker data format error: %s\n", err)
				continue
			}
			//删除配置
			key := *removeConfigReq.Key
			log.Printf("Receiver Routine: get deleted configure %s\n", key)
			controller.Remove(key)
		}
	}
}