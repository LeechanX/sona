package logic

import (
    "os"
    "net"
    "log"
    "fmt"
    "sync/atomic"
)

//现有连接个数
var numberOfConnections int32

func BrokerService() {
    tcpAddr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf("0.0.0.0:%d", GConf.BrokerPort))
    listen, err := net.ListenTCP("tcp", tcpAddr)
    if err != nil {
        log.Fatalf("%s\n", err)
        os.Exit(1)
    }
    log.Printf("create broker server(%s) successfully\n", fmt.Sprintf("0.0.0.0:%d", GConf.BrokerPort))
    defer listen.Close()

    for {
        conn, err := listen.AcceptTCP()
        if err != nil {
            log.Printf("%s\n", err)
            os.Exit(1)
        }
        //处理请求
        if atomic.LoadInt32(&numberOfConnections) < int32(GConf.AgentConnectionLimit) {
            CreateConnection(conn)
            log.Printf("current there are %d agent connections\n", numberOfConnections)
        } else {
            //直接关闭连接
            conn.Close()
            log.Println("connections is too much now")
        }
    }
}
