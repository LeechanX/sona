package logic

import "sync"

//被订阅列表结构
type SubscribedList struct {
	rwLock sync.RWMutex
	//serviceKey -> a set of Connections
	list map[string]map[*Connection]bool
}

//key被连接c所订阅
func (sl *SubscribedList) Subscribed(serviceKey string, c* Connection) {
	sl.rwLock.Lock()//写锁
	defer sl.rwLock.Unlock()

	sl.list[serviceKey][c] = true
}

func (sl *SubscribedList) RemoveKey(key string) {
	sl.rwLock.Lock()//写锁
	defer sl.rwLock.Unlock()
	delete(sl.list, key)
}

//连接c不再订阅key
func (sl *SubscribedList) UnSubscribed(key string, c* Connection) {
	sl.rwLock.Lock()//写锁
	defer sl.rwLock.Unlock()

	if _, ok := sl.list[key];ok {
		//remove
		delete(sl.list[key], c)
	}
}

//获取订阅了key的连接们
func (sl *SubscribedList) GetSubscribers(key string) []*Connection {
	sl.rwLock.RLock()
	defer sl.rwLock.RUnlock()
	//连接集合
	cs := make([]*Connection, 0)
	if _, ok := sl.list[key];ok {
		for c := range sl.list[key] {
			cs = append(cs, c)
		}
	}
	return cs
}

//被订阅列表
var Subscribed SubscribedList
