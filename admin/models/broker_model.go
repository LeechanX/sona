package models

import (
    "sona/common/net/tcp/client"
    "github.com/astaxie/beego"
    "github.com/golang/protobuf/proto"
    "sona/protocol"
    "time"
    "errors"
)

type ServiceConfig struct {
    ServiceKey string
    Version uint
    Confs map[string]string
}

func GetAdminClient() (*client.SyncClient, error) {
    ip := beego.AppConfig.String("brokeradminip")
    port, err := beego.AppConfig.Int("brokeradminport")
    if err != nil {
        beego.Error("broker admin port is wrong:", err)
        return nil, err
    }
    adminClient, err := client.CreateSyncClient(ip, port)
    if err != nil {
        beego.Error("create broker admin wrong:", err)
        return nil, err
    }
    return adminClient, nil
}

func Get(serviceKey string, cli *client.SyncClient) (*ServiceConfig, error) {
    req := &protocol.AdminGetConfigReq{}
    req.ServiceKey = proto.String(serviceKey)

    err := cli.Send(protocol.AdminGetConfigReqId, req)
    if err != nil {
        return nil, err
    }
    //接收 100ms超时
    timeout := 100 * time.Millisecond
    rsp := &protocol.AdminGetConfigRsp{}
    err = cli.Read(timeout, protocol.AdminGetConfigRspId, rsp)
    if err != nil {
        beego.Error("read AdminGetConfigRsp response:",err)
        return nil, err
    }
    if *rsp.Code == -1 {
        return nil, errors.New("this service is not exist")
    }

    serviceConf := &ServiceConfig{
        ServiceKey:*rsp.ServiceKey,
        Version:uint(*rsp.Version),
        Confs:make(map[string]string),
    }
    for i := 0;i < len(rsp.ConfKeys);i++ {
        key, value := rsp.ConfKeys[i], rsp.Values[i]
        serviceConf.Confs[key] = value
    }
    return serviceConf, nil
}

func Add(serviceKey string, client *client.SyncClient) error {
    //send add request
    req := &protocol.AdminAddConfigReq{}
    req.ServiceKey = proto.String(serviceKey)
    req.ConfKeys = make([]string, 0)
    req.Values = make([]string, 0)

    err := client.Send(protocol.AdminAddConfigReqId, req)
    if err != nil {
        return err
    }

    //接收 100ms超时
    timeout := 100 * time.Millisecond
    rsp := &protocol.AdminExecuteRsp{}
    err = client.Read(timeout, protocol.AdminExecuteRspId, rsp)
    if err != nil {
        return err
    }
    if *rsp.Code == 0 {
        return nil
    } else {
        return errors.New(*rsp.Error)
    }
}

func Update(serviceKey string, version uint, confs map[string]string, client *client.SyncClient) error {
    //build PB
    req := &protocol.AdminUpdConfigReq{}
    req.ServiceKey = proto.String(serviceKey)
    req.Version = proto.Uint32(uint32(version))
    req.ConfKeys = make([]string, 0)
    req.Values = make([]string, 0)

    for confKey, confValue := range confs {
        req.ConfKeys = append(req.ConfKeys, confKey)
        req.Values = append(req.Values, confValue)
    }
    err := client.Send(protocol.AdminUpdConfigReqId, req)
    if err != nil {
        return err
    }
    //接收 100ms超时
    timeout := 100 * time.Millisecond
    rsp := &protocol.AdminExecuteRsp{}
    err = client.Read(timeout, protocol.AdminExecuteRspId, rsp)
    if err != nil {
        return err
    }
    if *rsp.Code == 0 {
        return nil
    } else {
        return errors.New(*rsp.Error)
    }
}

