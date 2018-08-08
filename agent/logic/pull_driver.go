package logic

import (
    "log"
    "time"
    "sona/protocol"
    "github.com/golang/protobuf/proto"
)

//agent与broker重建连接后立刻拉取一次，以便告知broker：agent已订阅了哪些service key
func PullWhenStart() {
    serviceKeys := ConfController.GetAllServiceKeys()
    for serviceKey := range serviceKeys {
        log.Printf("pull serive %s directly\n", serviceKey)
        req := &protocol.PullServiceConfigReq{}
        req.ServiceKey = proto.String(serviceKey)
        req.Version = proto.Uint32(uint32(serviceKeys[serviceKey]))
        BrokerClient.Send(protocol.PullServiceConfigReqId, req)
    }
}

func PeriodPulling() {
    //周期性更新每个现有service的配置
    for {
        time.Sleep(time.Second * 100)
        serviceKeys := ConfController.GetAllServiceKeys()
        log.Printf("try to re-pull %d configures\n", len(serviceKeys))
        for serviceKey := range serviceKeys {
            req := &protocol.PullServiceConfigReq{}
            req.ServiceKey = proto.String(serviceKey)
            req.Version = proto.Uint32(uint32(serviceKeys[serviceKey]))
            BrokerClient.Send(protocol.PullServiceConfigReqId, req)
        }
        if len(serviceKeys) == 0 {
            time.Sleep(time.Second * 10)
        }
    }
}

