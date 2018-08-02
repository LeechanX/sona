package logic

import (
    "os"
    "log"
    "errors"
    "sona/protocol"
    "sona/broker/dao"
    "gopkg.in/mgo.v2"
    "sona/broker/conf"
    "sona/common/net/tcp"
    "github.com/golang/protobuf/proto"
)

//全局：admin server，服务于web操作
var AdminServer *tcp.Server

//错误
var (
    ErrEditByOther = errors.New("this service configure is editing by other user now")
    ErrExist = errors.New("this service is already exist")
    ErrVersionWrong = errors.New("version number is wrong")
    ErrNoChange = errors.New("no any different")
)

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

//新增配置
func AddConfig(serviceKey string, confKeys []string, values []string) error {
    //标注：serviceKey正在编辑中
    if !EditingControl.TryMarkEditing(serviceKey) {
        //说明正在被其他人编辑
        return ErrEditByOther
    }
    //编辑完成
    defer EditingControl.DoneEditing(serviceKey)

    //先在mongodb中查询是否有此serviceKey
    _, _, _, err := dao.GetDocument(serviceKey)
    if err == nil {
        //说明mongoDB中已经存在此service了，不允许执行
        return ErrExist
    }
    if err != mgo.ErrNotFound {
        //遇到了非"不存在"的错误
        return err
    }

    //新增的service版本号为1
    //则调用mongo更新接口
    err = dao.AddDocument(serviceKey, 1, confKeys, values)
    if err != nil {
        //调用出错，认为添加失败
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
        Version:proto.Uint32(1),
        ConfKeys:confKeys,
        Values:values,
    }
    for _, agent := range agents {
        agent.SendData(protocol.PushServiceConfigReqId, &pushReq)
    }
    //回写缓存
    CacheLayer.WriteBack(serviceKey, 1, confKeys, values)
    return nil
}

//修改配置
func UpdateConfig(serviceKey string, version uint, confKeys []string, values []string) error {
    //标注：serviceKey正在编辑中
    if !EditingControl.TryMarkEditing(serviceKey) {
        //说明正在被其他人编辑
        return ErrEditByOther
    }
    //编辑完成
    defer EditingControl.DoneEditing(serviceKey)

    //先在mongodb中查询此serviceKey
    originVersion, originConfKeys, originValues, err := dao.GetDocument(serviceKey)
    if err != nil {
        return err
    }
    //检查版本是否匹配
    if originVersion != version {
        return ErrVersionWrong
    }
    //检查是否有改动，防止无用请求
    if !isDifferent(originConfKeys, originValues, confKeys, values) {
        //没改动
        return ErrNoChange
    }

    newVersion := originVersion + 1
    err = dao.UpdateDocument(serviceKey, newVersion, confKeys, values)
    if err != nil {
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
        ConfKeys:confKeys,
        Values:values,
    }
    pushReq.Version = proto.Uint32(uint32(newVersion))
    for _, agent := range agents {
        agent.SendData(protocol.PushServiceConfigReqId, &pushReq)
    }
    //回写缓存
    CacheLayer.WriteBack(serviceKey, newVersion, confKeys, values)
    return nil
}

//消息ID与对应PB的映射
func adminMsgFactory(cmdId uint) proto.Message {
    switch cmdId {
    case protocol.AdminAddConfigReqId:
        return &protocol.AdminAddConfigReq{}
    case protocol.AdminCleanConfigReqId:
        return &protocol.AdminCleanConfigReq{}
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
        log.Printf("add configure %s meet error: %s\n", *req.ServiceKey, err)
        rsp.Code = proto.Int32(-1)
        rsp.Error = proto.String(err.Error())
    } else {
        log.Printf("add configure %s successfully\n", *req.ServiceKey)
        rsp.Code = proto.Int32(0)
        rsp.Error = proto.String("")
    }
    //回包
    session.SendData(protocol.AdminExecuteRspId, &rsp)
}

//AdminDelConfigReqId消息的回调函数
func cleanConfigHandler(session *tcp.Session, pb proto.Message) {
    log.Println("debug: into delete config callback")
    req := pb.(*protocol.AdminCleanConfigReq)
    //以空内容替换
    err := UpdateConfig(*req.ServiceKey, uint(*req.Version), []string{}, []string{})
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


//AdminUpdConfigReqId消息的回调函数
func updateConfigHandler(session *tcp.Session, pb proto.Message) {
    log.Println("debug: into update config callback")
    req := pb.(*protocol.AdminUpdConfigReq)
    rsp := protocol.AdminExecuteRsp{}

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
    //直接在缓存中读取
    confKeys, values, version := CacheLayer.GetData(*req.ServiceKey)
    if version == 0 {
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
    AdminServer.RegHandler(protocol.AdminCleanConfigReqId, cleanConfigHandler)
    AdminServer.RegHandler(protocol.AdminUpdConfigReqId, updateConfigHandler)
    AdminServer.RegHandler(protocol.AdminGetConfigReqId, getConfigHandler)
    //启动服务
    AdminServer.Start()
}
