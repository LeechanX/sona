package logic

import (
    "log"
    "sync"
    "time"
    "sona/core"
)

//周期性检查是否有待清理的serviceKey
type AccessRecord struct {
    lastUseTime map[string]int64
    mutex sync.Mutex
}

//agent启动加载共享内存后，使用共享内存中所有serviceKey，创建accessRecord
func GetAccessRecord(serviceKeys []string) *AccessRecord {
    record := &AccessRecord{
        lastUseTime:make(map[string]int64),
    }
    for _, serviceKey := range serviceKeys {
        record.Record(serviceKey)
    }
    return record
}

//client每次KeepUsing请求到来就更新一下使用记录
func (r *AccessRecord) Record(serviceKey string) {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    r.lastUseTime[serviceKey] = time.Now().Unix()
}

//删除近期未使用的serviceKey
func (r *AccessRecord) RemoveOutdated() []string {
    current := time.Now().Unix()
    r.mutex.Lock()
    defer r.mutex.Unlock()
    outdated := make([]string, 0)
    for serviceKey, ts := range r.lastUseTime {
        //获取一小时内都没使用的serviceKey
        if current - ts >= 3600 {
            outdated = append(outdated, serviceKey)
        }
    }
    //删除
    for _, serviceKey := range outdated {
        delete(r.lastUseTime, serviceKey)
    }
    return outdated
}

//周期性检查是否有需要删除的service key
func (r *AccessRecord) Cleaner(cc *core.ConfigController) {
    for {
        time.Sleep(time.Second * 10)
        outdated := r.RemoveOutdated()
        for _, serviceKey := range outdated {
            log.Printf("remove service %s because it hasn't used for a long time\n", serviceKey)
            cc.RemoveService(serviceKey)
        }
    }
}
