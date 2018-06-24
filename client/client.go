package client

import (
	"log"
	"net"
	"time"
	"errors"
	"easyconfig/core"
	"easyconfig/protocol"
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

func (client *EasyConfigClient) requestConfig(key string) error {
	//create body
	req := protocol.GetConfigReq{}
	req.Key = &key
	data := protocol.EncodeMessage(protocol.MsgTypeId_GetConfigReqId, &req)
	//发送
	_, err := client.conn.Write(data)
	return err
}

func (client *EasyConfigClient) Get(key string, timeout uint) (string, error) {
	if key == "" {
		return "", errors.New("empty key")
	}
	if timeout > 500 {
		timeout = 100
	}
	value, err := client.getter.Get(key)
	if err == nil || timeout == 0 {
		return value, err
	}

	//request to agent
	log.Printf("key %s is not exist, request agent\n", key)
	if udpErr := client.requestConfig(key); udpErr != nil {
		log.Printf("request agent error: %s\n", udpErr)
		return "", err
	}
	log.Printf("Wait for key %s\n", key)
	waitedTime := uint(0)
	for waitedTime < timeout {
		time.Sleep(time.Microsecond * 1)
		//重取一次
		value, err = client.getter.Get(key)
		if err == nil {
			return value, nil
		}
		waitedTime += 10
	}
	return "", err
}

func (client *EasyConfigClient) Close() {
	client.getter.Close()
	client.conn.Close()
}