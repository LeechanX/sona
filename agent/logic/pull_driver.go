package logic

import (
    "log"
    "time"
    "sync/atomic"
    "sona/core"
    "sona/protocol"
)

//agent与broker重建连接后立刻拉取一次，以便告知broker：agent已订阅了哪些service key
func PullWhenStart(controller *core.ConfigController, c *Connection) {
    serviceKeys := controller.GetAllServiceKeys()
    for serviceKey := range serviceKeys {
        if atomic.LoadInt32(&c.status) == kConnStatusConnected {
            req := protocol.PullServiceConfigReq{}
            req.ServiceKey = &serviceKey
            c.sendQueue<- &req
        }
    }
}

func PeriodicPull(controller *core.ConfigController, c *Connection) {
    //周期性更新每个现有service的配置
    for {
        time.Sleep(time.Second * 10)
        serviceKeys := controller.GetAllServiceKeys()
        for serviceKey := range serviceKeys {
            //sleep 1s
            log.Printf("Periodic Pull Routine: try to update %s's configures\n", serviceKey)
            if atomic.LoadInt32(&c.status) == kConnStatusConnected {
                req := protocol.PullServiceConfigReq{}
                req.ServiceKey = &serviceKey
                c.sendQueue<- &req
            } else {
                //连接已经断开，则不发送
                //DO NOTHING
            }
        }
        if len(serviceKeys) == 0 {
            time.Sleep(time.Second * 10)
        }
    }
}
