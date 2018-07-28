package logic

import (
    "sync"
    "sona/broker/dao"
)

//data:读多写少 key: string, value: string
//不使用sync.Map还是因为这个结构支持的方法太局限

type CacheConfigData struct {
    //格式：
    //serviceKey1: configKey1:configValue1, configKey2:configValue2...
    //serviceKey2: configKey1:configValue1, configKey2:configValue2...
    data map[string]*dao.ServiceData
    rwMutex sync.RWMutex
}

//全局配置
var CacheData CacheConfigData

//最初加载配置
func (cd *CacheConfigData) Load() error {
    newData, err := dao.ReloadAllData()
    if err != nil {
        return err
    }
    cd.rwMutex.Lock()
    cd.data = newData
    cd.rwMutex.Unlock()
    return nil
}

//读取缓存，获取serviceKey
func (cd *CacheConfigData) GetData(serviceKey string) ([]string, []string, uint) {
    cd.rwMutex.RLock()
    defer cd.rwMutex.RUnlock()
    data, ok := cd.data[serviceKey]
    if !ok {
        return []string{}, []string{}, 0
    }
    return data.ConfKeys, data.ConfValues, data.Version
}

//回写缓存
func (cd *CacheConfigData) WriteBack(serviceKey string, newVersion uint, configKeys []string, values []string) {
    serviceData := &dao.ServiceData{
        ServiceKey:serviceKey,
        Version:newVersion,
        ConfKeys:configKeys,
        ConfValues:values,
    }
    cd.rwMutex.Lock()
    defer cd.rwMutex.Unlock()
    if originServiceData, ok := cd.data[serviceKey];ok {
        if originServiceData.Version < newVersion {
            cd.data[serviceKey] = serviceData
        }
    } else {
        cd.data[serviceKey] = serviceData
    }
}
