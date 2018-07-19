package logic

import (
    "log"
    "sona/protocol"
    "sona/common/net/tcp/client"
    "github.com/golang/protobuf/proto"
)

//消息与PB的映射
func brokerClientMsgFactory(cmdId uint) proto.Message {
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
func SubscribeResultHandler(_ *client.AsyncClient, pb proto.Message) {
    log.Println("debug: subscribe result callback")
    req := pb.(*protocol.SubscribeBrokerRsp)
    if *req.Code == 0 {
        //订阅成功
        log.Printf("subscribe for %s successfully\n", *req.ServiceKey)
        ConfController.UpdateService(*req.ServiceKey, uint(*req.Version), req.ConfKeys, req.Values)
    } else {
        log.Printf("subscribe for %s failed\n", *req.ServiceKey)
    }
    rsp := &protocol.SubscribeAgentRsp{}
    rsp.ServiceKey = proto.String(*req.ServiceKey)
    rsp.Code = proto.Int32(*req.Code)
    //可能需要回复给biz
    //获取并删除UDP地址
    addrList := BizServer.SubscribeBook.GetSubscribers(*rsp.ServiceKey, true)
    for _, addr := range addrList {
        BizServer.Send(protocol.SubscribeAgentRspId, rsp, addr)
    }
}

//PushServiceConfigReqId消息的回调函数
func PushConfigHandler(_ *client.AsyncClient, pb proto.Message) {
    req := pb.(*protocol.PushServiceConfigReq)
    log.Printf("sona agent received the push request for service %s\n", *req.ServiceKey)
    //执行更新
    ConfController.UpdateService(*req.ServiceKey, uint(*req.Version), req.ConfKeys, req.Values)
}

//PullServiceConfigRspId消息的回调函数
func PullResultHandler(_ *client.AsyncClient, pb proto.Message) {
    req := pb.(*protocol.PullServiceConfigRsp)
    //执行更新
    ConfController.UpdateService(*req.ServiceKey, uint(*req.Version), req.ConfKeys, req.Values)
}

func CreateBrokerClient(ip string, port int, enableHeartbeat bool) *client.AsyncClient {
    cli := client.CreateAsyncClient(ip, port, enableHeartbeat)
    cli.SetFactory(brokerClientMsgFactory)
    //安装回调
    cli.RegHandler(protocol.SubscribeBrokerRspId, SubscribeResultHandler)
    cli.RegHandler(protocol.PushServiceConfigReqId, PushConfigHandler)
    cli.RegHandler(protocol.PullServiceConfigRspId, PullResultHandler)
    return cli
}