package logic

import "sync"

//被订阅列表结构
type SubscribedList struct {
    rwLock sync.RWMutex
    //serviceKey -> a set of Connections
    list map[string]map[*AgentConnection]bool
}

//key被连接c所订阅
func (sl *SubscribedList) Subscribed(serviceKey string, c* AgentConnection) {
    sl.rwLock.Lock()//写锁
    defer sl.rwLock.Unlock()
    sl.list[serviceKey][c] = true
}

//连接c不再订阅key
func (sl *SubscribedList) UnSubscribed(key string, c* AgentConnection) {
    sl.rwLock.Lock()//写锁
    defer sl.rwLock.Unlock()

    if _, ok := sl.list[key];ok {
        //remove
        delete(sl.list[key], c)
    }
}

//获取订阅了key的连接们
func (sl *SubscribedList) GetSubscribers(key string) []*AgentConnection {
    sl.rwLock.RLock()
    defer sl.rwLock.RUnlock()
    //连接集合
    cs := make([]*AgentConnection, 0)
    if _, ok := sl.list[key];ok {
        for c := range sl.list[key] {
            cs = append(cs, c)
        }
    }
    return cs
}

//被订阅列表
var SubscribedBook SubscribedList
