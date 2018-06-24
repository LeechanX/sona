package logic

import (
	"log"
	"time"
	"easyconfig/core"
	"easyconfig/protocol"
)

func PeriodicPull(controller *core.ConfigController, ch chan<- protocol.PullConfigReq) {
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
				ch<- pullConfigReq
			}
			//sleep 1s
			time.Sleep(time.Second * 1)
		}
	}
}
