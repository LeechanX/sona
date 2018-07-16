package udp

import (
    "sync"
    "net"
)

//被订阅列表结构
type SubscribeList struct {
    rwLock sync.RWMutex
    //info key -> a set of Address
    list map[interface{}][]*net.UDPAddr
}

//创建订阅列表
func CreateSubscribeList() *SubscribeList {
    sl := &SubscribeList{}
    sl.list = make(map[interface{}][]*net.UDPAddr)
    return sl
}

//key被某udp地址所订阅
func (sl *SubscribeList) Subscribe(infoKey interface{}, addr *net.UDPAddr) {
    sl.rwLock.Lock()//写锁
    defer sl.rwLock.Unlock()
    if _, ok := sl.list[infoKey];!ok {
        sl.list[infoKey] = make([]*net.UDPAddr, 0)
    }
    sl.list[infoKey] = append(sl.list[infoKey], addr)
}

//获取订阅了key的连接们
func (sl *SubscribeList) GetSubscribers(infoKey interface{}, del bool) []*net.UDPAddr {
    sl.rwLock.RLock()
    defer sl.rwLock.RUnlock()
    //连接集合
    if addrList, ok := sl.list[infoKey];ok {
        if del {
            delete(sl.list, infoKey)
        }
        return addrList
    }
    return nil
}

