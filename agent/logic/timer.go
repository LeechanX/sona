package logic

import (
	"log"
	"time"
	"sync/atomic"
	"easyconfig/core"
	"easyconfig/protocol"
)

func PeriodicPull(controller *core.ConfigController, c *Connection) {
	for {
		for idx := uint(0);idx < core.BucketCap;idx++ {
			confMap := controller.GetAll(idx)
			if len(confMap) != 0 {
				log.Printf("Periodic Pull Routine: try to update bucket %d\n", idx)
				pullConfigReq := protocol.PullConfigReq{}
				for key, value := range confMap {
					pullConfigReq.Keys = append(pullConfigReq.Keys, key)
					pullConfigReq.Values = append(pullConfigReq.Values, value)
				}
				if atomic.LoadUint32(&c.status) == CONNECTED {
					c.sendQueue<- pullConfigReq
				} else {
					//连接已经断开，则不发送
					//DO NOTHING
				}
			}
			//sleep 1s
			time.Sleep(time.Second * 1)
		}
	}
}
