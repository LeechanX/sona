package logic

import (
    "log"
    "time"
    "sona/broker/dao"
)

//周期性（30s）重加载所有数据
func PeriodReload() {
    for {
        time.Sleep(60 * time.Second)
        //先加载mongoDB数据
        newData, err := dao.ReloadAllData()
        if err != nil {
            log.Printf("load data from mongodb meet error: %s\n", err)
            continue
        }
        //筛选出可能要更新的数据
        candidate := make([]*dao.ServiceData, 0)
        for serviceKey, serviceData := range newData {
            _, _, version := CacheData.GetData(serviceKey)
            if version == 0 {
                //缓存中不存在
                candidate = append(candidate, serviceData)
            } else if version < serviceData.Version {
                //库中版本更大，数据more新
                candidate = append(candidate, serviceData)
            }
        }
        if len(candidate) > 0 {
            log.Println("after reload, database may have fresh data")
            log.Println("try to write fresh data to cache")
        }
        //尝试更新到缓存中
        for _, serviceData := range candidate {
            CacheData.WriteBack(serviceData.ServiceKey, serviceData.Version, serviceData.ConfKeys, serviceData.ConfValues)
        }
    }
}