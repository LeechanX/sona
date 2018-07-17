package logic

import (
    "log"
    "time"
    "sona/protocol"
)

//agent与broker重建连接后立刻拉取一次，以便告知broker：agent已订阅了哪些service key
func PullWhenStart() {
    serviceKeys := ConfController.GetAllServiceKeys()
    for serviceKey := range serviceKeys {
        req := &protocol.PullServiceConfigReq{}
        req.ServiceKey = &serviceKey
        *req.Version = uint32(serviceKeys[serviceKey])
        BrokerClient.Send(protocol.PullServiceConfigReqId, req)
    }
}

func PeriodPulling() {
    //周期性更新每个现有service的配置
    for {
        time.Sleep(time.Second * 10)
        serviceKeys := ConfController.GetAllServiceKeys()
        for serviceKey := range serviceKeys {
            log.Printf("Periodic Pull Routine: try to update %s's configures\n", serviceKey)
            req := &protocol.PullServiceConfigReq{}
            req.ServiceKey = &serviceKey
            *req.Version = uint32(serviceKeys[serviceKey])
            BrokerClient.Send(protocol.PullServiceConfigReqId, req)
        }
        if len(serviceKeys) == 0 {
            time.Sleep(time.Second * 10)
        }
    }
}

