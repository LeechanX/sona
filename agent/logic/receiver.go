package logic

import (
	"log"
	"easyconfig/core"
	"easyconfig/protocol"
	"github.com/golang/protobuf/proto"
)

func ReceiveFromBroker(controller *core.ConfigController, c *Connection) {
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
		//收到来自broker的回复
		if cmdId == protocol.MsgTypeId_PullConfigRspId {
			pullConfigRsp := protocol.PullConfigRsp{}
			err = proto.Unmarshal(pbData, &pullConfigRsp)
			if err != nil {
				log.Printf("receive from broker data format error: %s\n", err)
				return
			}

			if len(pullConfigRsp.Keys) != len(pullConfigRsp.Values) {
				continue
			}
			//依次更新每个结果
			for idx := range pullConfigRsp.Keys {
				key := pullConfigRsp.Keys[idx]
				log.Printf("get updated configure %s\n", key)
				value := pullConfigRsp.Values[idx]
				if value != "" {
					//说明是修改
					if err = controller.Set(key, value);err != nil {
						log.Panicf("%s\n", err)
					}
				} else {
					//说明需要删除
					controller.Remove(key)
				}
			}
		} else if cmdId == protocol.MsgTypeId_PushConfigReqId {
			//broker主动推送配置到来
			pushConfigRsp := protocol.PushConfigReq{}
			err = proto.Unmarshal(pbData, &pushConfigRsp)
			if err != nil {
				log.Printf("receive from broker data format error: %s\n", err)
				return
			}
			//更新配置
			key := *pushConfigRsp.Key
			log.Printf("Receiver Routine: get updated configure %s\n", key)
			value := *pushConfigRsp.Value
			if err = controller.Set(key, value);err != nil {
				log.Panicf("Set configure meet error: %s\n", err)
			}
		} else if cmdId == protocol.MsgTypeId_RemoveConfigReqId {
			//broker要求删除配置
			removeConfigReq := protocol.RemoveConfigReq{}
			err = proto.Unmarshal(pbData, &removeConfigReq)
			if err != nil {
				log.Printf("receive from broker data format error: %s\n", err)
				return
			}
			//删除配置
			key := *removeConfigReq.Key
			log.Printf("Receiver Routine: get deleted configure %s\n", key)
			controller.Remove(key)
		} else {
			log.Printf("unknown request cmd id: %d\n", cmdId)
		}
	}
}
