package logic

import (
    "log"
    "sync"
    "errors"
    "sona/broker/dao"
)

//data:读多写少 key: string, value: string
//不使用sync.Map还是因为这个结构支持的方法太局限

const (
    kStatusIdle = 1//空闲中
    kStatusEditing = 2//正在配置中
)

type ServiceData struct {
    version uint
    serviceKey string
    status int
    confKeys []string
    values []string
}

type ConfigureData struct {
    //格式：
    //serviceKey1: configKey1:configValue1, configKey2:configValue2...
    //serviceKey2: configKey1:configValue1, configKey2:configValue2...
    data map[string]*ServiceData
    rwMutex sync.RWMutex
}

//全局配置
var ConfigData ConfigureData

func (cfd *ConfigureData) Reset() error {
    dbDoc, err := dao.ReloadAllData()
    if err != nil {
        return err
    }
    newData := make(map[string]*ServiceData)
    //创新新数据
    for _, doc := range dbDoc {
        serviceData := &ServiceData{
            serviceKey:doc.ServiceKey,
            version:doc.Version,
            status:kStatusIdle,
            confKeys:doc.ConfKeys,
            values:doc.ConfValues,
        }
        newData[doc.ServiceKey] = serviceData
        log.Println("debug: service key: ", doc.ServiceKey)
        log.Println("debug: conf keys: ", doc.ConfKeys)
        log.Println("debug: conf values: ", doc.ConfValues)
    }
    ConfigData.rwMutex.Lock()
    cfd.data = newData
    ConfigData.rwMutex.Unlock()
    return nil
}

//新增配置操作
func (cfd *ConfigureData) AddConfig(serviceKey string, configKeys []string, values []string) (uint, error) {
    log.Println("debug: run add config operation")
    var version uint
    cfd.rwMutex.Lock()
    _, ok := cfd.data[serviceKey]
    if !ok {
        //在本地内存中先预先新增
        log.Println("debug: pre-save and set status = kStatusEditing")
        cfd.data[serviceKey] = &ServiceData{}
        cfd.data[serviceKey].version = 0
        cfd.data[serviceKey].serviceKey = serviceKey
        cfd.data[serviceKey].status = kStatusEditing
        cfd.rwMutex.Unlock()
    } else {
        version = cfd.data[serviceKey].version
        if len(cfd.data[serviceKey].confKeys) == 0 {
            //原有记录已被删除，可以被新增，检查是否在编辑中
            if cfd.data[serviceKey].status == kStatusEditing {
                log.Println("debug: another user is editing this configure")
                cfd.rwMutex.Unlock()
                return 0, errors.New("this service configure is in editing")
            } else {
                //标记为正在编辑
                log.Println("debug: mark kStatusEditing")
                cfd.data[serviceKey].status = kStatusEditing
                cfd.rwMutex.Unlock()
            }
        } else {
            log.Println("debug: data is exist")
            //已存在
            cfd.rwMutex.Unlock()
            return 0, errors.New("this service configure is already exist")
        }
    }
    if version != 0 {
        version += 1
    }
    //执行mongodb新增
    log.Println("debug: run add document in mongodb")
    err := dao.AddDocument(serviceKey, version, configKeys, values)
    cfd.rwMutex.Lock()
    defer cfd.rwMutex.Unlock()
    if err != nil {
        log.Println("debug: mongodb add error")
        //执行失败，回退
        //如果是刚添加的，则在内存中删除之
        if version != 0 {
            log.Println("debug: delete from broker memory")
            delete(cfd.data, serviceKey)
        } else {
            log.Println("debug: remark to kStatusIdle")
            //否则重置空闲状态
            cfd.data[serviceKey].status = kStatusIdle
        }
    } else {
        log.Println("debug: add ok and remark kStatusIdle")
        //执行成功
        cfd.data[serviceKey].confKeys = configKeys
        cfd.data[serviceKey].values = values
        //在内存中标记空闲
        cfd.data[serviceKey].status = kStatusIdle
    }
    return version, nil
}

//cas方式删除配置项
func (cfd *ConfigureData) DeleteData(serviceKey string, version uint) (uint, error) {
    return cfd.UpdateData(serviceKey, version, []string{}, []string{})
}

//cas方式修改配置项
func (cfd *ConfigureData) UpdateData(serviceKey string, version uint, configKeys []string, values []string) (uint, error) {
    cfd.rwMutex.Lock()
    _, ok := cfd.data[serviceKey]
    if ok {
        if cfd.data[serviceKey].version != version {
            //版本对不上号
            log.Println("debug: version is not equal")
            cfd.rwMutex.Unlock()
            return 0, errors.New("this service configure's version is wrong")
        } else {
            if cfd.data[serviceKey].status == kStatusEditing {
                //正在编辑中
                log.Println("debug: another user is editing this configure")
                cfd.rwMutex.Unlock()
                return 0, errors.New("this service configure is in editing")
            } else {
                //标记为正在编辑
                log.Println("debug: mark kStatusEditing")
                cfd.data[serviceKey].status = kStatusEditing
                cfd.rwMutex.Unlock()
            }
        }
    } else {
        //不存在
        log.Println("debug: not exist")
        cfd.rwMutex.Unlock()
        return 0, errors.New("this service configure is not exist")
    }

    //在mongodb中执行删除, 即把配置内容设置为空
    //版本+1
    version += 1
    log.Println("debug: run update document in mongodb")
    err := dao.UpdateDocument(serviceKey, version, configKeys, values)
    cfd.rwMutex.Lock()
    defer cfd.rwMutex.Unlock()
    if err == nil {
        log.Println("debug: mongodb update ok")
        //mongodb操作成功, 更新内存
        //更新版本
        cfd.data[serviceKey].version = version
        //将配置设置为空
        cfd.data[serviceKey].confKeys = configKeys
        cfd.data[serviceKey].values = values
    }
    log.Println("debug: mark kStatusIdle")
    //在内存中标记空闲
    cfd.data[serviceKey].status = kStatusIdle
    return version, nil
}

//获取serviceKey
func (cfd *ConfigureData) GetData(serviceKey string) ([]string, []string, uint) {
    cfd.rwMutex.RLock()
    defer cfd.rwMutex.RUnlock()
    data, ok := cfd.data[serviceKey]
    if !ok {
        return nil, nil, 0
    }
    return data.confKeys, data.values, data.version
}