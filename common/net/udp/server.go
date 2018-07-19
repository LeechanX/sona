package udp

import (
    "fmt"
    "net"
    "log"
    "time"
    "errors"
    "github.com/golang/protobuf/proto"
    "sona/common/net/protocol"
    "sync/atomic"
)

type SendTask struct {
    cmdId uint
    packet proto.Message
    addr *net.UDPAddr
}

const (
    Opening = iota
    Closed
)

type Server struct {
    Name string
    Ip string
    Port int
    status int32//1:open, 0 close
    conn *net.UDPConn
    sendQueue chan *SendTask
    SubscribeBook *SubscribeList//被订阅列表

    factory PacketFactory//消息ID与消息PB的映射函数
    callbacks map[uint]MsgHandler//消息回调
}

//消息ID与消息PB的映射函数类型
type PacketFactory func (uint) proto.Message
//遇到某消息ID的回调函数类型
type MsgHandler func (*Server, *net.UDPAddr, proto.Message)

//创建一个UDP服务
func CreateServer(serviceName string, ip string, port int) (*Server, error) {
    //创建UDP服务于client
    udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
    if err != nil {
        log.Printf("can's resolve udp address 127.0.0.1:9901\n")
        return nil, err
    }
    conn, err := net.ListenUDP("udp", udpAddr)
    if err != nil {
        log.Printf("error listening udp: %s\n", err)
        return nil, err
    }
    server := Server{
        Name:serviceName,
        Ip:ip,
        Port:port,
        status:Opening,
        SubscribeBook:CreateSubscribeList(),
    }
    server.conn = conn
    server.sendQueue = make(chan *SendTask, 100)
    server.callbacks = make(map[uint]MsgHandler)
    return &server, nil
}

//设置消息ID与消息PB的映射函数
func (server *Server) SetFactory(f PacketFactory) {
    server.factory = f
}

//设置消息ID对应的回调
func (server *Server) RegHandler(cmdId uint, handler MsgHandler) {
    server.callbacks[cmdId] = handler
}

//启动服务
func (server *Server) Start() error {
    if server.factory == nil {
        return errors.New("haven't set packet factory yet")
    }
    //启动发送G
    go server.sender()
    //作为接受G
    go server.receiver()
    return nil
}

//发消息
func (server *Server) Send(cmdId uint, pb proto.Message, addr *net.UDPAddr) bool {
    if atomic.LoadInt32(&server.status) == Opening {
        server.sendQueue <- &SendTask{
            cmdId:cmdId,
            packet:pb,
            addr:addr,
        }
        return true
    }
    return false
}

func (server *Server) Close() {
    if !atomic.CompareAndSwapInt32(&server.status, Opening, Closed) {
        //已关闭
        return
    }
    server.sendQueue<- nil
    server.conn.Close()
}

//收消息
func (server *Server) receiver() {
    defer server.Close()
    for {
        cmdId, addr, pbData, err := protocol.DecodeUDPMessage(server.conn)
        if err != nil {
            log.Printf("%s\n", err)
            continue
        }
        //doing
        handler, ok := server.callbacks[cmdId]
        if !ok {
            log.Printf("unknown request cmd id: %d\n", cmdId)
            continue
        }
        req := server.factory(cmdId)
        if req == nil {
            log.Printf("no pb mapping for cmd id: %d\n", cmdId)
            continue
        }
        if err := proto.Unmarshal(pbData, req);err != nil {
            log.Printf("server %s receive data format error: %s\n", server.Name, err)
            return
        }
        handler(server, addr, req)
    }
}

//回复消息
func (server *Server) sender() {
    defer server.Close()
    for {
        select {
        case task, ok := <-server.sendQueue:
            if !ok {
                //impossible code
                return
            }
            if task == nil {
                return
            }
            data := protocol.EncodeMessage(task.cmdId, task.packet)
            //超时设置
            server.conn.SetWriteDeadline(time.Now().Add(100*time.Millisecond))
            //忽略错误
            server.conn.WriteToUDP(data, task.addr)
        }
    }
}