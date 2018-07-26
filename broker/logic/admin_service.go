package logic

import (
    "os"
    "log"
    "sona/protocol"
    "sona/broker/conf"
    "sona/common/net/tcp"
    "github.com/golang/protobuf/proto"
)

//全局：admin server，服务于web操作
var AdminServer *tcp.Server

//新增配置
func AddConfig(serviceKey string, configKeys []string, values []string) error {
    newVersion, err := ConfigData.AddConfig(serviceKey, configKeys, values)
    if err != nil {
        log.Printf("add %s configure meet error: %s\n", serviceKey, err)
        return err
    }
    //如果有agent订阅, push给每个agent连接
    agents := BrokerServer.SubscribeBook.GetSubscribers(serviceKey)
    if len(agents) == 0 {
        return nil
    }
    //创建推送包
    pushReq := protocol.PushServiceConfigReq{
        ServiceKey:proto.String(serviceKey),
        Version:proto.Uint32(uint32(newVersion)),
        ConfKeys:configKeys,
        Values:values,
    }
    for _, agent := range agents {
        agent.SendData(protocol.PushServiceConfigReqId, &pushReq)
    }
    return nil
}

//修改配置
func UpdateConfig(serviceKey string, version uint, configKeys []string, values []string) error {
    newVersion, err := ConfigData.UpdateData(serviceKey, version, configKeys, values)
    if err != nil {
        log.Printf("update %s configure meet error: %s\n", serviceKey, err)
        return err
    }
    //如果有agent订阅, push给每个agent连接
    agents := BrokerServer.SubscribeBook.GetSubscribers(serviceKey)
    if len(agents) == 0 {
        return nil
    }
    log.Printf("debug: push updated data %s\n", serviceKey)
    //创建推送包
    pushReq := protocol.PushServiceConfigReq{
        ServiceKey:proto.String(serviceKey),
        Version:proto.Uint32(uint32(newVersion)),
        ConfKeys:configKeys,
        Values:values,
    }
    pushReq.Version = proto.Uint32(uint32(newVersion))
    for _, agent := range agents {
        agent.SendData(protocol.PushServiceConfigReqId, &pushReq)
    }
    return nil
}

//删除配置
func DelConfig(serviceKey string, version uint) error {
    newVersion, err := ConfigData.DeleteData(serviceKey, version)
    if err != nil {
        log.Printf("delete %s configure meet error: %s\n", serviceKey, err)
        return err
    }
    //如果有agent订阅, push给每个agent连接
    agents := BrokerServer.SubscribeBook.GetSubscribers(serviceKey)
    if len(agents) == 0 {
        return nil
    }
    log.Printf("debug: push deleted data %s\n", serviceKey)
    //创建推送包
    pushReq := protocol.PushServiceConfigReq{
        ServiceKey:proto.String(serviceKey),
        ConfKeys:[]string{},
        Values:[]string{},
    }
    pushReq.Version = proto.Uint32(uint32(newVersion))
    for _, agent := range agents {
        agent.SendData(protocol.PushServiceConfigReqId, &pushReq)
    }
    return nil
}

//消息ID与对应PB的映射
func adminMsgFactory(cmdId uint) proto.Message {
    switch cmdId {
    case protocol.AdminAddConfigReqId:
        return &protocol.AdminAddConfigReq{}
    case protocol.AdminDelConfigReqId:
        return &protocol.AdminDelConfigReq{}
    case protocol.AdminUpdConfigReqId:
        return &protocol.AdminUpdConfigReq{}
    case protocol.AdminGetConfigReqId:
        return &protocol.AdminGetConfigReq{}
    }
    return nil
}

//AdminAddConfigReqId消息的回调函数
func addConfigHandler(session *tcp.Session, pb proto.Message) {
    log.Println("debug: into add config callback")
    req := pb.(*protocol.AdminAddConfigReq)

    err := AddConfig(*req.ServiceKey, req.ConfKeys, req.Values)
    rsp := protocol.AdminExecuteRsp{}
    if err != nil {
        rsp.Code = proto.Int32(-1)
        rsp.Error = proto.String(err.Error())
    } else {
        rsp.Code = proto.Int32(0)
        rsp.Error = proto.String("")
    }
    //回包
    session.SendData(protocol.AdminExecuteRspId, &rsp)
}

//AdminDelConfigReqId消息的回调函数
func delConfigHandler(session *tcp.Session, pb proto.Message) {
    log.Println("debug: into delete config callback")
    req := pb.(*protocol.AdminDelConfigReq)

    err := DelConfig(*req.ServiceKey, uint(*req.Version))
    rsp := protocol.AdminExecuteRsp{}
    if err != nil {
        rsp.Code = proto.Int32(-1)
        rsp.Error = proto.String(err.Error())
    } else {
        rsp.Code = proto.Int32(0)
        rsp.Error = proto.String("")
    }
    //回包
    session.SendData(protocol.AdminExecuteRspId, &rsp)
}

//检查两组配置是否有区别
func isDifferent(k1 []string, v1 []string, k2 []string, v2 []string) bool {
    if len(k1) != len(k2) {
        return true//显然不同
    }
    kv1 := make(map[string]string)
    kv2 := make(map[string]string)
    for i := 0;i < len(k1);i++ {
        k, v := k1[i], v1[i]
        kv1[k] = v
    }
    for i := 0;i < len(k2);i++ {
        k, v := k2[i], v2[i]
        kv2[k] = v
    }
    for k := range kv1 {
        if _, ok := kv2[k];ok {
            if kv1[k] != kv2[k] {
                return true
            }
        } else {
            return true
        }
    }
    return false
}

//AdminUpdConfigReqId消息的回调函数
func updateConfigHandler(session *tcp.Session, pb proto.Message) {
    log.Println("debug: into update config callback")
    req := pb.(*protocol.AdminUpdConfigReq)
    rsp := protocol.AdminExecuteRsp{}
    //在mongoDB中执行
    originKeys, originValues, version := ConfigData.GetData(*req.ServiceKey)
    if version != uint(*req.Version) {
        rsp.Code = proto.Int32(-1)
        rsp.Error = proto.String("this service configure's version is wrong")
        //回包
        session.SendData(protocol.AdminExecuteRspId, &rsp)
        return
    }
    //检查是否有改动
    if !isDifferent(originKeys, originValues, req.ConfKeys, req.Values) {
        rsp.Code = proto.Int32(-1)
        rsp.Error = proto.String("no any changed")
        //回包
        session.SendData(protocol.AdminExecuteRspId, &rsp)
        return
    }

    err := UpdateConfig(*req.ServiceKey, uint(*req.Version), req.ConfKeys, req.Values)

    if err != nil {
        rsp.Code = proto.Int32(-1)
        rsp.Error = proto.String(err.Error())
    } else {
        rsp.Code = proto.Int32(0)
        rsp.Error = proto.String("")
    }
    //回包
    session.SendData(protocol.AdminExecuteRspId, &rsp)
}

//AdminGetConfigReqId消息的回调函数
func getConfigHandler(session *tcp.Session, pb proto.Message) {
    log.Println("debug: into get config callback")
    req := pb.(*protocol.AdminGetConfigReq)
    rsp := protocol.AdminGetConfigRsp{}
    rsp.ServiceKey = proto.String(*req.ServiceKey)
    confKeys, values, version := ConfigData.GetData(*req.ServiceKey)
    if confKeys == nil {
        //不存在
        rsp.Code = proto.Int32(-1)
        rsp.Version = proto.Uint32(0)
    } else {
        rsp.Code = proto.Int32(0)
        rsp.Version = proto.Uint32(uint32(version))
        rsp.ConfKeys = confKeys
        rsp.Values = values
    }
    //回包
    session.SendData(protocol.AdminGetConfigRspId, &rsp)
}

func StartAdminService() {
    server, err := tcp.CreateServer("admin", "0.0.0.0", conf.GlobalConf.AdminPort, uint32(conf.GlobalConf.AdminConnectionLimit))
    if err != nil {
        log.Println(err)
        os.Exit(1)
    }
    AdminServer = server
    //注册消息ID与PB的映射
    AdminServer.SetFactory(adminMsgFactory)
    //注册所有回调
    AdminServer.RegHandler(protocol.AdminAddConfigReqId, addConfigHandler)
    AdminServer.RegHandler(protocol.AdminDelConfigReqId, delConfigHandler)
    AdminServer.RegHandler(protocol.AdminUpdConfigReqId, updateConfigHandler)
    AdminServer.RegHandler(protocol.AdminGetConfigReqId, getConfigHandler)
    //启动服务
    AdminServer.Start()
}
