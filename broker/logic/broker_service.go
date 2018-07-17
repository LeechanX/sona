package logic

import (
    "os"
    "log"
    "sona/protocol"
    "sona/broker/conf"
    "sona/common/net/tcp"
    "github.com/golang/protobuf/proto"
)

//全局：broker server，服务于agent
var BrokerServer *tcp.Server

//消息ID与对应PB的映射
func brokerMapping(cmdId uint) proto.Message {
    switch cmdId {
    case protocol.SubscribeReqId:
        return &protocol.SubscribeReq{}
    case protocol.PullServiceConfigReqId:
        return &protocol.PullServiceConfigReq{}
    }
    return nil
}

//SubscribeReqId消息的回调函数
func SubscribeHandler(session *tcp.Session, pb proto.Message) {
    req, ok := pb.(*protocol.SubscribeReq)
    if !ok {
        log.Println("get SubscribeReq pb error")
        return
    }
    //订阅：此连接对*req.ServiceKey感兴趣
    session.Subscribe(*req.ServiceKey)
    //创建回包
    rsp := protocol.SubscribeBrokerRsp{}
    rsp.ServiceKey = req.ServiceKey
    //查看是否有此配置
    keys, values, version := ConfigData.GetData(*req.ServiceKey)
    if keys == nil {
        *rsp.Code = -1//订阅失败
    } else {
        *rsp.Code = 0//订阅成功
        //填充配置
        *rsp.Version = uint32(version)
        rsp.ConfKeys = keys
        rsp.Values = values
    }
    //回包
    session.SendData(protocol.SubscribeBrokerRspId, &rsp)
}

//PullServiceConfigReqId消息的回调函数
func PullConfigHandler(session *tcp.Session, pb proto.Message) {
    req, ok := pb.(*protocol.PullServiceConfigReq)
    if !ok {
        log.Println("get SubscribeReq pb error")
        return
    }
    //订阅：此连接对*req.ServiceKey感兴趣
    session.Subscribe(*req.ServiceKey)
    //创建回包
    rsp := protocol.PullServiceConfigRsp{}
    rsp.ServiceKey = req.ServiceKey

    //查看是否有此配置 (必然有)
    keys, values, version := ConfigData.GetData(*req.ServiceKey)
    *rsp.Version = uint32(version)
    if version > uint(*req.Version) {
        //agent端的版本过时了
        rsp.ConfKeys = keys
        rsp.Values = values
    }
    //回包
    session.SendData(protocol.PullServiceConfigRspId, &rsp)
}

func StartBrokerService() {
    server, err := tcp.CreateServer("broker", "0.0.0.0", conf.GlobalConf.BrokerPort, uint32(conf.GlobalConf.BrokerConnectionLimit))
    if err != nil {
        log.Println(err)
        os.Exit(1)
    }
    BrokerServer = server
    //注册消息ID与PB的映射
    BrokerServer.SetMapping(brokerMapping)
    //注册所有回调
    BrokerServer.RegHandler(protocol.SubscribeReqId, SubscribeHandler)
    BrokerServer.RegHandler(protocol.PullServiceConfigReqId, PullConfigHandler)
    //启动服务
    BrokerServer.Start()
}
