package logic

import (
    "sync"
    "sona/broker/dao"
    "log"
    "time"
    "gopkg.in/mgo.v2"
)

//data:读多写少 key: string, value: string
//不使用sync.Map还是因为这个结构支持的方法太局限

//缓存单元，记录了加载时间
type CacheUint struct {
    dao.ServiceData
    loadTime int64
}

//缓存层
type CacheLayerStructure struct {
    data map[string]*CacheUint
    rwMutex sync.RWMutex
}

//全局配置
var CacheLayer CacheLayerStructure

func init() {
    CacheLayer.data = make(map[string]*CacheUint)
}

//读取缓存，获取serviceKey, 最后一个返回值是版本号，是否存在系统中不存在则为0
func (cd *CacheLayerStructure) GetData(serviceKey string) ([]string, []string, uint) {
    cd.rwMutex.RLock()
    data, ok := cd.data[serviceKey]
    cd.rwMutex.RUnlock()
    if ok {
        return data.ConfKeys, data.ConfValues, data.Version
    }

    version, confKeys, values, err := dao.GetDocument(serviceKey)
    if err == nil {
        //存在, 写回到缓存里
        cd.WriteBack(serviceKey, version, confKeys, values)
        return confKeys, values, version
    }

    if err != mgo.ErrNotFound {
        log.Printf("get data from mongo db meet error: %s\n", err)
    }
    return []string{}, []string{}, 0
}

//回写缓存
func (cd *CacheLayerStructure) WriteBack(serviceKey string, newVersion uint, configKeys []string, values []string) {
    unit := &CacheUint{
        loadTime:time.Now().Unix(),
    }
    unit.ServiceKey = serviceKey
    unit.Version = newVersion
    unit.ConfKeys = configKeys
    unit.ConfValues = values

    cd.rwMutex.Lock()
    defer cd.rwMutex.Unlock()
    if originUnit, ok := cd.data[serviceKey];ok {
        if originUnit.Version < newVersion {
            log.Printf("write back to cache for service %s\n", serviceKey)
            cd.data[serviceKey] = unit
        }
    } else {
        log.Printf("write back to cache for service %s\n", serviceKey)
        cd.data[serviceKey] = unit
    }
}

//清理过期缓存 (过期时间：delay，秒)
func (cd *CacheLayerStructure) ClearExpired(delay int64) {
    for {
        time.Sleep(time.Second)
        currentTs := time.Now().Unix()
        //先获取过期service key
        expired := make([]string, 0)
        cd.rwMutex.RLock()
        for serviceKey, unit := range cd.data {
            if currentTs-unit.loadTime >= delay {
                expired = append(expired, serviceKey)
            }
        }

        cd.rwMutex.RUnlock()
        //对每个可能过期key，二次检查时间并尝试删除
        cd.rwMutex.Lock()
        for _, serviceKey := range expired {
            if unit, ok := cd.data[serviceKey]; ok {
                if currentTs-unit.loadTime >= delay {
                    //执行删除
                    delete(cd.data, serviceKey)
                }
            }
        }
        cd.rwMutex.Unlock()
    }
}
