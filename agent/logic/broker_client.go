package logic

import (
    "log"
    "sona/protocol"
    "sona/common/net/tcp/client"
    "github.com/golang/protobuf/proto"
)

//消息与PB的映射
func brokerClientMapping(cmdId uint) proto.Message {
    switch cmdId {
    case protocol.SubscribeBrokerRspId:
        return &protocol.SubscribeBrokerRsp{}
    case protocol.PushServiceConfigReqId:
        return &protocol.PushServiceConfigReq{}
    case protocol.PullServiceConfigRspId:
        return &protocol.PullServiceConfigRsp{}
    }
    return nil
}

//SubscribeBrokerRspId消息的回调函数
func SubscribeResultHandler(_ *client.Client, pb proto.Message) {
    req, ok := pb.(*protocol.SubscribeBrokerRsp)
    if !ok {
        log.Println("get SubscribeBrokerRsp pb error")
        return
    }
    if *req.Code == 0 {
        //订阅成功
        ConfController.UpdateService(*req.ServiceKey, uint(*req.Version), req.ConfKeys, req.Values)
    }
    rsp := &protocol.SubscribeAgentRsp{}
    rsp.ServiceKey = req.ServiceKey
    rsp.Code = req.Code
    //可能需要回复给biz
    //获取并删除UDP地址
    addrList := BizServer.SubscribeBook.GetSubscribers(*rsp.ServiceKey, true)
    for _, addr := range addrList {
        BizServer.Send(protocol.SubscribeAgentRspId, rsp, addr)
    }
}

//PushServiceConfigReqId消息的回调函数
func PushConfigHandler(_ *client.Client, pb proto.Message) {
    req, ok := pb.(*protocol.PushServiceConfigReq)
    if !ok {
        log.Println("get PushServiceConfigReq pb error")
        return
    }
    //执行更新
    ConfController.UpdateService(*req.ServiceKey, uint(*req.Version), req.ConfKeys, req.Values)
}

//PullServiceConfigRspId消息的回调函数
func PullResultHandler(_ *client.Client, pb proto.Message) {
    req, ok := pb.(*protocol.PullServiceConfigRsp)
    if !ok {
        log.Println("get PullServiceConfigRsp pb error")
        return
    }
    //执行更新
    ConfController.UpdateService(*req.ServiceKey, uint(*req.Version), req.ConfKeys, req.Values)
}

func CreateBrokerClient(ip string, port int) *client.Client {
    client := client.CreateClient(ip, port)
    client.SetMapping(brokerClientMapping)
    //安装回调
    client.RegHandler(protocol.SubscribeBrokerRspId, SubscribeResultHandler)
    client.RegHandler(protocol.PushServiceConfigReqId, PushConfigHandler)
    client.RegHandler(protocol.PullServiceConfigRspId, PullResultHandler)
    return client
}