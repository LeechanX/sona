package client

import (
    "fmt"
    "net"
    "log"
    "sync"
    "time"
    "errors"
    "sync/atomic"
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
type PBMapping func (uint) proto.Message
//遇到某消息ID的回调函数类型
type MsgHandler func (*AsyncClient, proto.Message)

type AsyncClient struct {
    Ip string
    Port int
    conn *net.TCPConn
    status int32
    wg sync.WaitGroup//用于等待连接的读G、写G退出，标明连接被关闭
    sendQueue chan *SendTask

    mapping PBMapping//消息ID与消息PB的映射函数
    hooks map[uint]MsgHandler//消息回调
}

//创建一个client结构体
func CreateAsyncClient(ip string, port int) *AsyncClient {
    return &AsyncClient{
        Ip:ip,
        Port:port,
        conn:nil,
        status:kConnStatusDisconnected,
        sendQueue:make(chan *SendTask, 1000),
        mapping:nil,
        hooks:make(map[uint]MsgHandler),
    }
}

//设置消息ID与消息PB的映射函数
func (c *AsyncClient) SetMapping(m PBMapping) {
    c.mapping = m
}

//设置消息ID对应的回调
func (c *AsyncClient) RegHandler(cmdId uint, handler MsgHandler) {
    c.hooks[cmdId] = handler
}

//执行连接
func (c *AsyncClient) Connect() error {
    if c.mapping == nil {
        return errors.New("haven't set pb mapping yet")
    }
    if atomic.LoadInt32(&c.status) == kConnStatusConnected {
        return errors.New("already built the connection")
    }
    //创建TCP客户端去连接broker
    tcpAddr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.Ip, c.Port))
    log.Printf("connecting to %s\n", fmt.Sprintf("%s:%d", c.Ip, c.Port))
    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        log.Printf("can's connect tcp address %s\n", fmt.Sprintf("%s:%d", c.Ip, c.Port))
        return err
    }
    log.Printf("connected to broker %s successfully\n", fmt.Sprintf("%s:%d", c.Ip, c.Port))
    c.conn = conn
    //设置状态为已连接
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
}

//发送消息
func (c *AsyncClient) Send(cmdId uint, pb proto.Message) bool {
    if atomic.LoadInt32(&c.status) == kConnStatusConnected {
        c.sendQueue<- &SendTask{
            cmdId:cmdId,
            packet:pb,
        }
        return true
    }
    return false
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
        if err != nil {
            log.Printf("%s\n", err)
            return
        }
        req := c.mapping(cmdId)
        if req == nil {
            log.Printf("no pb mapping for cmd id: %d\n", cmdId)
            continue
        }
        if err := proto.Unmarshal(pbData, req);err != nil {
            log.Printf("client receive data format error: %s\n", err)
            return
        }

        handler, ok := c.hooks[cmdId]
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