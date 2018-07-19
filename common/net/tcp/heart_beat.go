package tcp

import (
    "sync"
    "time"
    "sync/atomic"
    "sona/common/net/protocol"
    "github.com/golang/protobuf/proto"
)

var HeartbeatReqId uint = 666
var HeartbeatRspId uint = 777

const (
    HeartbeatPeriod = 30//心跳周期，30s
    LostThreshold = 150//失联阈值，150s
)

type ActiveList struct {
    mutex sync.Mutex//没必要读写锁，因为多写一读
    sessions map[*Session]bool//当前连接总集合
}

//新建连接就放入此集合
func (list *ActiveList) AddSession(s *Session) {
    list.mutex.Lock()
    defer list.mutex.Unlock()
    if _, ok := list.sessions[s];ok {
        panic("impossible code: connection is exist!")
    }
    list.sessions[s] = true
}

//连接关闭则从此集合中删除
func (list *ActiveList) RemoveSession(s *Session) {
    list.mutex.Lock()
    defer list.mutex.Unlock()
    if _, ok := list.sessions[s];!ok {
        panic("impossible code: connection is not exist!")
    }
    delete(list.sessions, s)
}

//收到心跳回包，激活连接: 更新收包时间
func (list *ActiveList) Activate(s *Session) {
    list.mutex.Lock()
    defer list.mutex.Unlock()
    if _, ok := list.sessions[s];ok {
        currentTs := time.Now().Unix()
        atomic.StoreInt64(&s.HeartBeatRspTs, currentTs)//原子更新
    }
}

//获取需要再次发送心跳的连接而发心跳、认为已经不再活跃的连接而回收连接
func (list *ActiveList) HeartbeatProbe() {
    probe := make([]*Session, 0)//需要再次发送心跳的连接
    lost := make([]*Session, 0)//已经不再活跃的连接
    currentTs := time.Now().Unix()
    list.mutex.Lock()
    for s := range list.sessions {
        heartBeatRspTs := atomic.LoadInt64(&s.HeartBeatRspTs)
        if currentTs - heartBeatRspTs >= LostThreshold {
            //已很久没用回包了 认为不再活跃
            lost = append(lost, s)
        } else if currentTs - s.HeartBeatReqTs >= HeartbeatPeriod {
            //否则，如果又该发心跳了
            probe = append(probe, s)
        }
    }
    list.mutex.Unlock()
    //关闭非活跃连接
    for _, session := range lost {
        session.Close()
    }
    //发心跳
    req := &protocol.HeartbeatReq{
        Useless:proto.Bool(true),
    }
    for _, session := range probe {
        if session.SendData(HeartbeatReqId, req) {
            //设置最新发送心跳时间
            session.HeartBeatReqTs = currentTs
        }
    }
}

//收到HeartBeatReqId的回调函数：
//回复心跳
func HeartbeatReqHandler(session *Session, _ proto.Message) {
    rsp := &protocol.HeartbeatRsp{
        Useless:proto.Bool(true),
    }
    session.SendData(HeartbeatRspId, rsp)
}

//收到HeartbeatRspId的回调函数：
//记录时间
func HeartbeatRspHandler(session *Session, _ proto.Message) {
    //激活此连接
    session.server.actives.Activate(session)
}
