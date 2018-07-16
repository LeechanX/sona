package tcp

import "sync"

//被订阅列表结构
type SubscribeList struct {
    rwLock sync.RWMutex
    //info key -> a set of Connections
    list map[interface{}]map[*Session]bool
}

//创建订阅列表
func CreateSubscribeList() *SubscribeList {
    sl := &SubscribeList{}
    sl.list = make(map[interface{}]map[*Session]bool)
    return sl
}

//key被连接session所订阅
func (sl *SubscribeList) Subscribe(infoKey interface{}, session *Session) {
    sl.rwLock.Lock()//写锁
    defer sl.rwLock.Unlock()
    if _, ok := sl.list[infoKey];!ok {
        sl.list[infoKey] = make(map[*Session]bool)
    }
    sl.list[infoKey][session] = true
}

//连接session不再订阅key
func (sl *SubscribeList) UnSubscribe(infoKey interface{}, session *Session) {
    sl.rwLock.Lock()//写锁
    defer sl.rwLock.Unlock()
    if _, ok := sl.list[infoKey];ok {
        delete(sl.list[infoKey], session)
    }
    if len(sl.list[infoKey]) == 0 {
        delete(sl.list, infoKey)
    }
}

//获取订阅了key的连接们
func (sl *SubscribeList) GetSubscribers(infoKey interface{}) []*Session {
    sl.rwLock.RLock()
    defer sl.rwLock.RUnlock()
    //连接集合
    sessionList := make([]*Session, 0)
    if _, ok := sl.list[infoKey];ok {
        for session := range sl.list[infoKey] {
            sessionList = append(sessionList, session)
        }
    }
    return sessionList
}

