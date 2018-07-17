package tcp

import (
    "net"
    "log"
    "time"
    "sync/atomic"
    "sona/common/net/protocol"
    "github.com/golang/protobuf/proto"
)

//记录当前网络状态
const (
    kConnStatusConnected = int32(1)//正常连接
    kConnStatusDisconnected//已断开
)

//待发送数据
type SendTask struct {
    cmdId uint
    packet proto.Message
}

type Session struct {
    status int32
    conn *net.TCPConn
    //发送数据
    sendQueue chan *SendTask
    //此连接的订阅信息（用于推送）
    subscribeList map[interface{}]bool
    //指向tcp服务入口
    server *Server
}

//创建会话
func CreateSession(server *Server, c *net.TCPConn) {
    session := Session{
        status:kConnStatusConnected,
        conn:c,
        sendQueue:make(chan *SendTask, 1000),
        subscribeList:make(map[interface{}]bool),
        server:server,
    }
    //当前连接数+1
    atomic.AddInt32(&session.server.NumberOfConnections, 1)
    //启动发送G
    go session.sender()
    //启动接收G
    go session.receiver()
}

//此会话订阅了某信息
func (session *Session) Subscribe(infoKey interface{}) {
    session.subscribeList[infoKey] = true
    //在全局订阅列表上也更新
    session.server.SubscribeBook.Subscribe(infoKey, session)
}

//连接是否存活
func (session *Session) IsClosed() bool {
    return atomic.LoadInt32(&session.status) == kConnStatusDisconnected
}

//关闭连接
func (session *Session) Close() {
    if !atomic.CompareAndSwapInt32(&session.status, kConnStatusConnected,
         kConnStatusDisconnected) {
        //已被关闭过
        return
    }

    //需要在全局被订阅列表里删除每个关联
    for infoKey := range session.subscribeList {
        session.server.SubscribeBook.UnSubscribe(infoKey, session)
    }
    //为防止写channel产生panic，不关闭channel，仅发nil
    session.sendQueue<- nil
    session.conn.Close()
    atomic.AddInt32(&session.server.NumberOfConnections, -1)
}

//发送消息
func (session *Session) SendData(cmdId uint, pb proto.Message) bool {
    if !session.IsClosed() {
        session.sendQueue<- &SendTask{
            cmdId:cmdId,
            packet:pb,
        }
        return true
    }
    return false
}

//接收消息的goroutine
func (session *Session) receiver() {
    //可能是网络出错，于是调用CloseConnect会主动关闭连接
    //也可能是其他G关闭了连接，这时调用Close将什么也不干
    defer session.Close()
    for {
        cmdId, pbData, err := protocol.DecodeTCPMessage(session.conn)
        if err != nil {
            log.Printf("%s\n", err)
            return
        }

        handler, ok := session.server.hooks[cmdId]
        if !ok {
            log.Printf("unknown request cmd id: %d\n", cmdId)
            continue
        }
        req := session.server.mapping(cmdId)
        if req == nil {
            log.Printf("no pb mapping for cmd id: %d\n", cmdId)
            continue
        }
        if err := proto.Unmarshal(pbData, req);err != nil {
            log.Printf("server %s receive data format error: %s\n", session.server.Name, err)
            return
        }
        handler(session, req)
    }
}

//发送消息的goroutine
func (session *Session) sender() {
    //可能是网络出错，于是调用CloseConnect会主动关闭连接
    //也可能是其他G关闭了连接，这时调用CloseConnect将什么也不干
    defer session.Close()
    for {
        select {
        case task, ok := <- session.sendQueue:
            if !ok {
                return//impossible code: 已经关闭了
            } else if task == nil {
                return//连接已经关闭
            }
            data := protocol.EncodeMessage(task.cmdId, task.packet)
            session.conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
            if _, err := session.conn.Write(data);err != nil {
                log.Printf("send data error: %s\n", err)
                return
            }
        }
    }
}