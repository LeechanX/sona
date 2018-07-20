package tcp

import (
    "net"
    "log"
    "fmt"
    "errors"
    "sync/atomic"
    "github.com/golang/protobuf/proto"
    "time"
)

//消息ID与消息PB的映射函数类型，根据消息ID给出对应的PB包
type PacketFactory func (uint) proto.Message
//遇到某消息ID的回调函数类型
type MsgHandler func (*Session, proto.Message)

type Server struct {
    Name string
    Ip string
    Port int
    MaxConnectionNumber uint32//最大连接个数
    NumberOfConnections int32//目前连接个数，hold by atomic
    SubscribeBook *SubscribeList//被订阅列表

    listen *net.TCPListener
    factory PacketFactory//消息ID与消息PB的映射函数
    callbacks map[uint]MsgHandler//消息回调
    actives *ActiveList//活跃列表:心跳维护
}

//https://www.cnblogs.com/concurrency/p/4043271.html
//创建TCP服务 参数：服务名,IP,PORT,最大连接个数
func CreateServer(serviceName string, ip string, port int, maxConnectionNumber uint32) (*Server, error) {
    tcpAddr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", ip, port))
    listen, err := net.ListenTCP("tcp", tcpAddr)
    if err != nil {
        log.Printf("%s\n", err)
        return nil, err
    }
    server := &Server{
        Name:serviceName,
        Ip:ip,
        Port:port,
        MaxConnectionNumber:maxConnectionNumber,
        NumberOfConnections:0,
        SubscribeBook:CreateSubscribeList(),
        listen:listen,
        factory:nil,
        callbacks:make(map[uint]MsgHandler),
        actives:CreateActiveList(),
    }
    //先主动注册收到心跳的回调
    server.callbacks[HeartbeatReqId] = HeartbeatReqHandler
    log.Printf("create %s server(%s) successfully\n", serviceName, fmt.Sprintf("%s:%d", ip, port))
    return server, nil
}

//打开发心跳机制
func (server *Server) EnableHeartbeat() {
    log.Println("open hearbeat probe mechanism")
    //注册收到心跳回复的回调
    server.callbacks[HeartbeatRspId] = HeartbeatRspHandler
    //开启一个G用于驱动心跳检测
    go func(s *Server) {
        for {
            s.actives.HeartbeatProbe()
            time.Sleep(time.Second)
        }
    }(server)
}

//设置消息ID与消息PB的映射函数
func (server *Server) SetFactory(f PacketFactory) {
    server.factory = f
}

//设置消息ID对应的回调
func (server *Server) RegHandler(cmdId uint, handler MsgHandler) {
    if cmdId == HeartbeatReqId || cmdId == HeartbeatRspId {
        log.Printf("can't use cmdId value %d or %d\n", HeartbeatReqId, HeartbeatRspId)
        return
    }
    server.callbacks[cmdId] = handler
}

//启动服务
func (server *Server) Start() error {
    if server.factory == nil {
        return errors.New("haven't set packet factory yet")
    }
    log.Printf("start %s server(%s) serivce\n", server.Name, fmt.Sprintf("%s:%d", server.Ip, server.Port))
    defer server.listen.Close()

    for {
        conn, err := server.listen.AcceptTCP()
        if err != nil {
            log.Printf("tcp service %s listen error: %s\n", server.Name, err)
            return err
        }
        //处理请求
        if atomic.LoadInt32(&server.NumberOfConnections) < int32(server.MaxConnectionNumber) {
            CreateSession(server, conn)
            log.Printf("current there are %d connections for service %s\n",
                server.NumberOfConnections, server.Name)
        } else {
            //直接关闭连接
            conn.Close()
            log.Println("connections is too much now")
        }
    }
}
