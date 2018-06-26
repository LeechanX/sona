package logic

import (
	"sync"
	"errors"
)

//data:读多写少 key: string, value: string
//不使用sync.Map还是因为这个结构函数太单调
//var DataMap sync.Map

type ConfiguresData struct {
	data map[string]string
	rwMutex sync.RWMutex
}

var ConfigData ConfiguresData

//cas方式新增、修改配置
func (cfd *ConfiguresData) AddOrUpdateData(key string, oldValue string, newValue string) error {
	cfd.rwMutex.Lock()

	//是否需要push
	var needPush bool
	value, ok := cfd.data[key]
	if oldValue == value {
		//可以设置
		cfd.data[key] = newValue
		//是修改，需要下发
		if ok {
			needPush = true
		}
		cfd.rwMutex.Unlock()
	} else {
		//值不同，不可下发
		cfd.rwMutex.Unlock()
		return errors.New("please retry, cas wrong")
	}

	if needPush {
		//获取目标agent
		targets := Subscribed.GetSubscribers(key)
		for _, target := range targets {
			target.PushUpdatedData(key, value)
		}
	}
	return nil
}

//cas方式删除配置，需要Push
func (cfd *ConfiguresData) DeleteData(key string, oldValue string) error {
	cfd.rwMutex.Lock()

	value, ok := cfd.data[key]
	if !ok {
		cfd.rwMutex.Unlock()
		return errors.New("no exist")
	}
	if value != oldValue {
		cfd.rwMutex.Unlock()
		return errors.New("please retry, cas wrong")
	}
	delete(cfd.data, key)
	cfd.rwMutex.Unlock()

	//获取目标agent
	targets := Subscribed.GetSubscribers(key)
	//删除key
	Subscribed.RemoveKey(key)
	for _, target := range targets {
		target.PushDeletedData(key)
	}
	return nil
}

//获取配置
func (cfd *ConfiguresData) GetData(key string) string {
	cfd.rwMutex.RLock()
	defer cfd.rwMutex.RUnlock()
	if value, ok := cfd.data[key];ok {
		return value
	}
	return ""
}
