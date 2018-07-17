package logic

import (
    "net"
    "sona/protocol"
    "sona/common/net/udp"
    "github.com/golang/protobuf/proto"
)

//KeepUsingReqId消息的回调
func KeepUsingHandler(_ *udp.Server, _ *net.UDPAddr, pb proto.Message) {
    //业务上报心跳
    req := pb.(*protocol.KeepUsingReq)
    //如果agent不存在此serviceKey，那么去broker上尝试拉取
    if !ConfController.IsServiceExist(*req.ServiceKey) {
        //发送给broker拉取goroutine：拉取此配置
        pullReq := protocol.PullServiceConfigReq{}
        pullReq.ServiceKey = proto.String(*req.ServiceKey)
        pullReq.Version = proto.Uint32(0)

        BrokerClient.Send(protocol.PullServiceConfigReqId, &pullReq)
    } else {
        //否则更新本地时间戳
        AccessRecordTable.Record(*req.ServiceKey)
    }
}

//SubscribeReqId消息的回调
func SubscribeReqHandler(server *udp.Server, addr *net.UDPAddr, pb proto.Message) {
    req := pb.(*protocol.SubscribeReq)
    //client向agent订阅配置
    if ConfController.IsServiceExist(*req.ServiceKey) {
        //本地已经有了，则回复
        rsp := &protocol.SubscribeAgentRsp{}
        rsp.Code = proto.Int32(0)
        rsp.ServiceKey = proto.String(*req.ServiceKey)
        server.Send(protocol.SubscribeAgentRspId, rsp, addr)
    } else {
        //否则发送给broker拉取goroutine：去Broker订阅此配置
        BrokerClient.Send(protocol.SubscribeReqId, req)
        //记录addr订阅serviceKey，以便当serviceKey订阅信息返回时，可以准确推送给对应的UDP客户端
        server.SubscribeBook.Subscribe(*req.ServiceKey, addr)
    }
}

//消息ID与PB的映射关系
func agentMapping(cmdId uint) proto.Message {
    switch cmdId {
    case protocol.KeepUsingReqId:
        return &protocol.KeepUsingReq{}
    case protocol.SubscribeReqId:
        return &protocol.SubscribeReq{}
    }
    return nil
}

//创建一个面向biz服务的UDP服务
func CreateBizServer(ip string, port int) (*udp.Server, error) {
    server, err := udp.CreateServer("sona-agent", ip, port)
    if err != nil {
        return nil, err
    }
    server.SetMapping(agentMapping)
    //安装回调
    server.RegHandler(protocol.KeepUsingReqId, KeepUsingHandler)
    server.RegHandler(protocol.SubscribeReqId, SubscribeReqHandler)
    return server, nil
}
