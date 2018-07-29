package main

import (
    "os"
    "fmt"
    "time"
    "strconv"
    "sona/common/net/tcp"
    "sona/common/net/tcp/client"
    "sona/common/net/protocol"
    "github.com/golang/protobuf/proto"
)

func main() {
    if len(os.Args) != 3 {
        fmt.Println("usage: ./broker_detect broker_ip broker_port")
        os.Exit(0)//ignore
    }
    brokerIp := os.Args[1]
    brokerPort, err := strconv.Atoi(os.Args[2])
    if err != nil {
        fmt.Println("usage: ./broker_detect broker_ip broker_port")
        os.Exit(0)//ignore
    }
    detect, err := client.CreateSyncClient(brokerIp, brokerPort)
    if err != nil {
        //连接失败
        fmt.Printf("detect error: %s\n", err)
        os.Exit(1)
    }
    err = detect.Send(tcp.HeartbeatReqId, &protocol.HeartbeatReq{
        Useless:proto.Bool(true),
    })
    if err != nil {
        //发送失败
        fmt.Printf("detect error: %s\n", err)
        os.Exit(1)
    }
    rsp := &protocol.HeartbeatRsp{}
    err = detect.Read(100 * time.Millisecond, tcp.HeartbeatRspId, rsp)
    if err != nil {
        //接收失败
        fmt.Printf("detect error: %s\n", err)
        os.Exit(1)
    }
    fmt.Println("detect successfully")
    os.Exit(0)
}