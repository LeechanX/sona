package logic

import (
    "sync"
    "errors"
    "sona/broker/dao"
)

//data:读多写少 key: string, value: string
//不使用sync.Map还是因为这个结构支持的方法太局限

type ServiceData struct {
    conf map[string]string
    version uint
    serviceKey string
}

type ConfigureData struct {
    //格式：
    //serviceKey1: configKey1:configValue1, configKey2:configValue2...
    //serviceKey2: configKey1:configValue1, configKey2:configValue2...
    data map[string]ServiceData
    rwMutex sync.RWMutex
}

//全局配置
var ConfigData ConfigureData

func (cfd *ConfigureData) Reset(dbDoc []*dao.ConfigureDocument) {
    newData := make(map[string]*ServiceData)
    //创新新数据
    for _, doc := range dbDoc {
        serviceData := &ServiceData{}
        serviceData.serviceKey = doc.ServiceKey
        serviceData.version = doc.Version
        serviceData.conf = make(map[string]string)
        length := len(doc.ConfKeys)
        for i := 0;i < length;i++ {
            k, v := doc.ConfKeys[i], doc.ConfValues[i]
            serviceData.conf[k] = v
        }
    }
    //TODO 创建与重置之间有新的更改怎么办。。。。。。
    //TODO 且主broker并没有必要读DB更新
    //TODO 难点：Broker如何知道主从切换，即如何知道自己成了主、成了备
    //重置

}

//cas方式新增、修改配置
//返回值：bool表示是否需要推送
func (cfd *ConfigureData) AddOrUpdateData(serviceKey string, configKey string, oldValue string, newValue string) (bool, error) {
    cfd.rwMutex.Lock()
    defer cfd.rwMutex.Unlock()
    var needPush bool
    if _, ok := cfd.data[serviceKey];!ok {
        //不存在，添加serviceKey
        cfd.data[serviceKey] = make(map[string]string)
        //检查CAS
        if oldValue != "" {
            return false, errors.New("please retry, cas wrong")
        }
        //新创建的serviceKey，所以不需要推送，不可能有人已订阅
    } else {
        originValue := cfd.data[serviceKey][configKey]
        //检查CAS
        if oldValue != originValue {
            return false, errors.New("please retry, cas wrong")
        }
        needPush = true
    }
    //添加，需要push
    cfd.data[serviceKey][configKey] = newValue
    return needPush, nil
}

//cas方式删除配置项，必须要重推
func (cfd *ConfigureData) DeleteData(serviceKey string, configKey string, oldValue string) error {
    cfd.rwMutex.Lock()
    defer cfd.rwMutex.Unlock()
    var err error = nil
    if _, ok := cfd.data[serviceKey];ok {
        if originValue, ok := cfd.data[serviceKey][configKey];ok {
            //检查cas
            if originValue != oldValue {
                err = errors.New("please retry, cas wrong")
            } else {
                delete(cfd.data[serviceKey], configKey)
                if len(cfd.data[serviceKey]) == 0 {
                    delete(cfd.data, serviceKey)
                }
            }
        }
    }
    return err
}

//获取配置
func (cfd *ConfigureData) GetData(serviceKey string, configKey string) string {
    cfd.rwMutex.RLock()
    defer cfd.rwMutex.RUnlock()
    if _, ok := cfd.data[serviceKey];ok {
        if value, ok := cfd.data[serviceKey][configKey];ok {
            return value
        }
    }
    return ""
}

//查看是否存在serviceKey
func (cfd *ConfigureData) IsServiceKeyExist(serviceKey string) map[string]string {
    cfd.rwMutex.RLock()
    defer cfd.rwMutex.RUnlock()
    return cfd.data[serviceKey]
}