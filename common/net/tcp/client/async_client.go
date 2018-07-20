package client

import (
    "fmt"
    "net"
    "log"
    "sync"
    "time"
    "errors"
    "sync/atomic"
    "sona/common/net/tcp"
    "sona/common/net/protocol"
    "github.com/golang/protobuf/proto"
)

const (
    kConnStatusConnected = iota
    kConnStatusDisconnected
)

//待发送数据
type SendTask struct {
    cmdId uint
    packet proto.Message
}

//消息ID与消息PB的映射函数类型
type PacketFactory func (uint) proto.Message
//遇到某消息ID的回调函数类型
type MsgHandler func (*AsyncClient, proto.Message)

type AsyncClient struct {
    Ip string
    Port int
    conn *net.TCPConn
    status int32
    wg sync.WaitGroup//用于等待连接的读G、写G退出，标明连接被关闭
    sendQueue chan *SendTask

    factory PacketFactory//消息ID与消息PB的映射函数
    callbacks map[uint]MsgHandler//消息回调
    HeartBeatRspTs int64//上次收到心跳回复的时间戳，两个G使用，故原子操作
}

//创建一个client结构体
func CreateAsyncClient(ip string, port int, enableHeartBeat bool) *AsyncClient {
    cli := &AsyncClient{
        Ip:ip,
        Port:port,
        conn:nil,
        status:kConnStatusDisconnected,
        sendQueue:make(chan *SendTask, 1000),
        factory:nil,
        callbacks:make(map[uint]MsgHandler),
    }
    //主动注册：收到心跳请求的回调
    cli.callbacks[tcp.HeartbeatReqId] = func (c *AsyncClient, _ proto.Message) {
        rsp := &protocol.HeartbeatRsp{
            Useless:proto.Bool(true),
        }
        c.Send(tcp.HeartbeatRspId, rsp)
    }
    if enableHeartBeat {
        currentTs := time.Now().Unix()
        atomic.StoreInt64(&cli.HeartBeatRspTs, currentTs)
        //注册收到心跳回复的回调
        cli.callbacks[tcp.HeartbeatRspId] = func(c *AsyncClient, _ proto.Message) {
            //更新时间
            atomic.StoreInt64(&c.HeartBeatRspTs, time.Now().Unix())
        }
        //开启心跳检测G
        go cli.HeartbeatProbe()
    }

    return cli
}

//心跳检测G
func (c *AsyncClient) HeartbeatProbe() {
    var lastSendTs int64
    req := &protocol.HeartbeatReq{
        Useless:proto.Bool(true),
    }
    for {
        time.Sleep(time.Second)
        if c.isConnected() {
            //建立了连接且开启了心跳检测
            currentTs := time.Now().Unix()
            if currentTs - atomic.LoadInt64(&c.HeartBeatRspTs) >= tcp.LostThreshold {
                //连接不再活跃，关闭
                log.Println("connection is inactive too long, so close it")
                c.close()
            } else if currentTs - lastSendTs >= tcp.HeartbeatPeriod {
                //发心跳
                c.Send(tcp.HeartbeatReqId, req)
            }
        }
    }
}

//设置消息ID与消息PB的映射函数
func (c *AsyncClient) SetFactory(f PacketFactory) {
    c.factory = f
}

//设置消息ID对应的回调
func (c *AsyncClient) RegHandler(cmdId uint, handler MsgHandler) {
    c.callbacks[cmdId] = handler
}

//执行连接
func (c *AsyncClient) Connect() error {
    if c.factory == nil {
        return errors.New("haven't set packet factory yet")
    }
    if atomic.LoadInt32(&c.status) == kConnStatusConnected {
        return errors.New("already built the connection")
    }
    //创建TCP客户端去连接broker
    tcpAddr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.Ip, c.Port))
    log.Printf("connecting to %s\n", fmt.Sprintf("%s:%d", c.Ip, c.Port))
    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        log.Printf("can't connect tcp address %s\n", fmt.Sprintf("%s:%d", c.Ip, c.Port))
        return err
    }
    log.Printf("connected to %s successfully\n", fmt.Sprintf("%s:%d", c.Ip, c.Port))
    c.conn = conn
    //设置状态为已连接
    currentTs := time.Now().Unix()
    atomic.StoreInt64(&c.HeartBeatRspTs, currentTs)
    atomic.StoreInt32(&c.status, kConnStatusConnected)
    //创建两个协程：读、写
    c.wg.Add(1)
    go c.receiver()
    c.wg.Add(1)
    go c.sender()
    return nil
}

//等待连接关闭
func (c *AsyncClient) Wait() {
    c.wg.Wait()
    log.Printf("connection with %s:%d is closed\n", c.Ip, c.Port)
}

//发送消息
func (c *AsyncClient) Send(cmdId uint, pb proto.Message) bool {
    if c.isConnected() {
        c.sendQueue<- &SendTask{
            cmdId:cmdId,
            packet:pb,
        }
        return true
    }
    return false
}

//连接是否建立
func (c *AsyncClient) isConnected() bool {
    return atomic.LoadInt32(&c.status) == kConnStatusConnected
}

//关闭连接
func (c *AsyncClient) close() {
    if !atomic.CompareAndSwapInt32(&c.status, kConnStatusConnected, kConnStatusDisconnected) {
        log.Println("already closed the connection")
        return
    }
    log.Println("now close the connection")
    //防止panic，不关闭管道，而是发送消息nil告知已无消息
    c.sendQueue<- nil
    //关闭连接
    c.conn.Close()
}

//读取消息
func (c *AsyncClient) receiver() {
    defer func() {
        //可能是网络出错，于是调用CloseConnect会主动关闭连接
        //也可能是其他G关闭了连接，这时调用Close将什么也不干
        c.close()
        c.wg.Done()
    }()
    for {
        cmdId, pbData, err := protocol.DecodeTCPMessage(c.conn)
        var req proto.Message
        if err != nil {
            log.Printf("%s\n", err)
            return
        }
        if cmdId == tcp.HeartbeatReqId {
            //远端发来heartbeat
            req = &protocol.HeartbeatReq{}
        } else if cmdId == tcp.HeartbeatRspId {
            //远端回复heartbeat应答
            req = &protocol.HeartbeatRsp{}
        } else {
            req = c.factory(cmdId)
        }
        if req == nil {
            log.Printf("no packet factory for cmd id: %d\n", cmdId)
            continue
        }
        if err := proto.Unmarshal(pbData, req);err != nil {
            log.Printf("client receive data format error: %s\n", err)
            return
        }

        handler, ok := c.callbacks[cmdId]
        if !ok {
            log.Printf("unknown request cmd id: %d\n", cmdId)
            continue
        }
        handler(c, req)
    }
}

//发消息
func (c *AsyncClient) sender() {
    defer func() {
        //可能是网络出错，于是调用CloseConnect会主动关闭连接
        //也可能是其他G关闭了连接，这时调用CloseConnect将什么也不干
        c.close()
        c.wg.Done()
    }()
    for {
        select {
        case task, ok := <- c.sendQueue:
            if !ok {
                //impossible code
                return
            } else if task == nil {
                //说明连接已经关闭
                return
            }
            data := protocol.EncodeMessage(task.cmdId, task.packet)
            //设置100ms的超时
            c.conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
            if _, err := c.conn.Write(data);err != nil {
                log.Printf("send data error: %s\n", err)
                return
            }
        }
    }

}