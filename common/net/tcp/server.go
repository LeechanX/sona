package tcp

import (
    "net"
    "log"
    "fmt"
    "errors"
    "sync/atomic"
    "github.com/golang/protobuf/proto"
)

//消息ID与消息PB的映射函数类型
type PBMapping func (uint) proto.Message
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
    mapping PBMapping//消息ID与消息PB的映射函数
    hooks map[uint]MsgHandler//消息回调
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
        mapping:nil,
        hooks:make(map[uint]MsgHandler),
    }

    log.Printf("create %s server(%s) successfully\n", serviceName, fmt.Sprintf("%s:%d", ip, port))
    return server, nil
}

//设置消息ID与消息PB的映射函数
func (server *Server) SetMapping(m PBMapping) {
    server.mapping = m
}

//设置消息ID对应的回调
func (server *Server) RegHandler(cmdId uint, handler MsgHandler) {
    server.hooks[cmdId] = handler
}

//启动服务
func (server *Server) Start() error {
    if server.mapping == nil {
        return errors.New("haven't set pb mapping yet")
    }

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
