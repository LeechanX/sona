package api

import (
    "time"
    "errors"
    "strings"
    "sona/core"
    "sona/protocol"
    "sona/common/net/udp/client"
)

type SonaApi struct {
    getter *core.ConfigGetter
    serviceKey string
    udpClient *client.Client
}

func GetApi(serviceKey string) (*SonaApi, error) {
    if !core.IsValidityServiceKey(serviceKey) {
        return nil, errors.New("not a valid sona service key")
    }
    api := SonaApi{}
    getter, err := core.GetConfigGetter()
    if err != nil {
        return nil, err
    }
    //create UDP client
    udpClient, err := client.CreateClient("127.0.0.1", 9901)
    if err != nil {
        getter.Close()
        return nil, err
    }
    api.getter = getter
    api.udpClient = udpClient
    api.serviceKey = serviceKey
    api.subscribe(serviceKey)
    //启动一个保活routine:告诉agent我一直在使用此serviceKey
    go api.keepUsing()
    return &api, nil
}

func (api *SonaApi) keepUsing() {
    req := &protocol.KeepUsingReq{ServiceKey:&api.serviceKey}
    for {
        time.Sleep(time.Second * 10)
        //tell agent
        api.udpClient.Send(protocol.KeepUsingReqId, req)
    }
}

func (api *SonaApi) subscribe(serviceKey string) error {
    if !core.IsValidityServiceKey(serviceKey) {
        return errors.New("service key format error")
    }
    //发起订阅消息
    req := &protocol.SubscribeReq{}
    req.ServiceKey = &serviceKey
    err := api.udpClient.Send(protocol.SubscribeReqId, req)
    if err != nil {
        return err
    }

    rsp := protocol.SubscribeAgentRsp{}
    //接收 300ms超时
    timeout := 300 * time.Millisecond
    err = api.udpClient.Read(timeout, protocol.SubscribeAgentRspId, &rsp)
    if err != nil {
        return err
    }
    //收到包，处理
    if *rsp.ServiceKey != serviceKey {
        return errors.New("udp receive error data")
    }
    //订阅失败
    if *rsp.Code == -1 {
        return errors.New("no such a service in system right now")
    }
    return nil
}

//获取value
func (api *SonaApi) Get(section string, key string) string {
    confKey := section + "." + key
    if !core.IsValidityConfKey(confKey) {
        return ""
    }
    return api.getter.Get(api.serviceKey, confKey)
}

//获取value并以列表解析
func (api *SonaApi) GetList(section string, key string) []string {
    confKey := section + "." + key
    if !core.IsValidityConfKey(confKey) {
        return nil
    }
    value := api.getter.Get(api.serviceKey, confKey)
    items := strings.Split(value, ",")
    for idx, item := range items {
        items[idx] = strings.TrimSpace(item)
    }
    return items
}

func (api *SonaApi) Close() {
    api.getter.Close()
    api.udpClient.Close()
}