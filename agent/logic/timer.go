package logic

import (
	"log"
	"time"
	"sync/atomic"
	"easyconfig/core"
	"easyconfig/protocol"
)

func PeriodicPull(controller *core.ConfigController, c *Connection) {
	//周期性更新每个现有service的配置
	for {
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
			time.Sleep(time.Second * 10)
		}
		if len(serviceKeys) == 0 {
			time.Sleep(time.Second * 10)
		}
	}
}
