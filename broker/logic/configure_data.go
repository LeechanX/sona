package logic

import (
	"sync"
	"errors"
)

//data:读多写少 key: string, value: string
//不使用sync.Map还是因为这个结构函数太单调
//var DataMap sync.Map

type ConfiguresData struct {
	//格式：
	//serviceKey1: configKey1:configValue1, configKey2:configValue2...
	//serviceKey2: configKey1:configValue1, configKey2:configValue2...
	data map[string]map[string]string
	rwMutex sync.RWMutex
}

//全局配置
var ConfigData ConfiguresData

//cas方式新增、修改配置
//返回值：bool表示是否需要推送
func (cfd *ConfiguresData) AddOrUpdateData(serviceKey string, configKey string, oldValue string, newValue string) (bool, error) {
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
func (cfd *ConfiguresData) DeleteData(serviceKey string, configKey string, oldValue string) error {
	cfd.rwMutex.Lock()
	defer cfd.rwMutex.Unlock()
	var err error = nil
	if _, ok := cfd.data[serviceKey];ok {
		if originValue, ok := cfd.data[serviceKey][configKey];ok {
			//检查cas
			if originValue != oldValue {
				err = errors.New("please retry, cas wrong")
			}
		}
	}
	return err
}

//获取配置
func (cfd *ConfiguresData) GetData(serviceKey string, configKey string) string {
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
func (cfd *ConfiguresData) IsServiceKeyExist(serviceKey string) map[string]string {
	cfd.rwMutex.RLock()
	defer cfd.rwMutex.RUnlock()
	return cfd.data[serviceKey]
}