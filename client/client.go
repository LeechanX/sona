package client

import (
	"log"
	"net"
	"time"
	"errors"
	"easyconfig/core"
	"easyconfig/protocol"
	"github.com/golang/protobuf/proto"
	"strings"
)

type EasyConfigClient struct {
	getter *core.ConfigGetter
	conn *net.UDPConn
}

func GetClient() (*EasyConfigClient, error) {
	client := EasyConfigClient{}
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
	return &client, nil
}

func (client *EasyConfigClient) Subscribe(serviceKey string) error {
	if !core.IsValidityServiceKey(serviceKey) {
		return errors.New("service key format error")
	}
	//发起订阅消息
	req := protocol.SubscribeReq{}
	req.ServiceKey = &serviceKey
	data := protocol.EncodeMessage(protocol.MsgTypeId_SubscribeReqId, &req)
	//发送 50ms超时
	client.conn.SetWriteDeadline(time.Now().Add(50 * time.Millisecond))
	_, err := client.conn.Write(data)
	if err != nil {
		return err
	}
	//接收 300ms超时
	client.conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	cmdId, _, pbData, err := protocol.DecodeUDPMessage(client.conn)
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
func (client *EasyConfigClient) Get(key string) (string, error) {
	if !core.IsValidityKey(key) {
		return "", errors.New("empty key format error")
	}
	return client.getter.Get(key)
}

//获取value并以列表解析
func (client *EasyConfigClient) GetList(key string) ([]string, error) {
	if !core.IsValidityKey(key) {
		return nil, errors.New("empty key format error")
	}
	value, err := client.getter.Get(key)
	if err != nil {
		return nil, err
	}
	items := strings.Split(value, ",")
	for idx, item := range items {
		items[idx] = strings.TrimSpace(item)
	}
	return items, nil
}

func (client *EasyConfigClient) Close() {
	client.getter.Close()
	client.conn.Close()
}