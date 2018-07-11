package client

import (
    "log"
    "net"
    "time"
    "errors"
    "strings"
    "sona/core"
    "sona/protocol"
    "github.com/golang/protobuf/proto"
)

type SonaClient struct {
    getter *core.ConfigGetter
    serviceKey string
    conn *net.UDPConn
}

func GetClient(serviceKey string) (*SonaClient, error) {
    if !core.IsValidityServiceKey(serviceKey) {
        return nil, errors.New("not a valid sona service key")
    }
    client := SonaClient{}
    getter, err := core.GetConfigGetter()
    if err != nil {
        log.Panicf("get config getter error: %s\n", err)
        return nil, err
    }
    //create UDP client
    remoteAddr := net.UDPAddr{
        IP:   net.IPv4(127, 0, 0, 1),
        Port: 9901,
    }
    conn, err := net.DialUDP("udp", nil, &remoteAddr)
    if err != nil {
        log.Panicf("dial udp error: %s\n", err)
        getter.Close()
        return nil, err
    }
    client.getter = getter
    client.conn = conn
    client.serviceKey = serviceKey
    client.subscribe(serviceKey)
    //启动一个保活routine:告诉agent我一直在使用此serviceKey
    client.keepUsing()
    return &client, nil
}

func (c *SonaClient) keepUsing() {
    req := protocol.KeepUsingReq{ServiceKey:&c.serviceKey}
    data := protocol.EncodeMessage(protocol.MsgTypeId_KeepUsingReqId, &req)
    for {
        time.Sleep(time.Second * 10)
        //tell agent
        c.conn.SetWriteDeadline(time.Now().Add(50 * time.Millisecond))
        c.conn.Write(data)
    }
}

func (c *SonaClient) subscribe(serviceKey string) error {
    if !core.IsValidityServiceKey(serviceKey) {
        return errors.New("service key format error")
    }
    //发起订阅消息
    req := protocol.SubscribeReq{}
    req.ServiceKey = &serviceKey
    data := protocol.EncodeMessage(protocol.MsgTypeId_SubscribeReqId, &req)
    //发送 50ms超时
    c.conn.SetWriteDeadline(time.Now().Add(50 * time.Millisecond))
    _, err := c.conn.Write(data)
    if err != nil {
        return err
    }
    //接收 300ms超时
    c.conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
    cmdId, _, pbData, err := protocol.DecodeUDPMessage(c.conn)
    if err != nil {
        return err
    }
    if cmdId != protocol.MsgTypeId_SubscribeAgentRspId {
        return errors.New("udp receive error data")
    }
    //收到包
    rsp := protocol.SubscribeAgentRsp{}
    if err = proto.Unmarshal(pbData, &rsp);err != nil {
        return err
    }
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
func (c *SonaClient) Get(section string, key string) string {
    confKey := section + "." + key
    if !core.IsValidityConfKey(confKey) {
        return ""
    }
    return c.getter.Get(c.serviceKey, confKey)
}

//获取value并以列表解析
func (c *SonaClient) GetList(section string, key string) []string {
    confKey := section + "." + key
    if !core.IsValidityConfKey(confKey) {
        return nil
    }
    value := c.getter.Get(c.serviceKey, confKey)
    items := strings.Split(value, ",")
    for idx, item := range items {
        items[idx] = strings.TrimSpace(item)
    }
    return items
}

func (c *SonaClient) Close() {
    c.getter.Close()
    c.conn.Close()
}