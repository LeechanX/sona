package logic

import (
	"os"
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

//新增配置
func AddConfig(serviceKey string, configKey string, value string) error {
	needPush, err := ConfigData.AddOrUpdateData(serviceKey, configKey, "", value)
	if err != nil {
		log.Printf("Add config meet error: %s\n", err)
		return err
	}
	if needPush {
		//push给每个agent连接
		agents := SubscribedBook.GetSubscribers(serviceKey)
		for _, agent := range agents {
			agent.PushAddOrUpdated(serviceKey, configKey, value)
		}
	}
	return nil
}

//修改配置
func UpdateConfig(serviceKey string, configKey string, oldValue string, newValue string) error {
	if oldValue == newValue {
		return nil
	}
	needPush, err := ConfigData.AddOrUpdateData(serviceKey, configKey, oldValue, newValue)
	if err != nil {
		log.Printf("Update config meet error: %s\n", err)
		return err
	}
	if needPush {
		//push给每个agent连接
		agents := SubscribedBook.GetSubscribers(serviceKey)
		for _, agent := range agents {
			agent.PushAddOrUpdated(serviceKey, configKey, newValue)
		}
	}
	return nil
}

//删除配置
func DelConfig(serviceKey string, configKey string, oldValue string) error {
	if err := ConfigData.DeleteData(serviceKey, configKey, oldValue);err != nil {
		log.Printf("Delete config meet error: %s\n", err)
		return err
	}
	//push给每个agent连接
	agents := SubscribedBook.GetSubscribers(serviceKey)
	for _, agent := range agents {
		agent.PushDeleted(serviceKey, configKey)
	}
	return nil
}

func AdminService() {
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf("0.0.0.0:%d", GConf.AdminPort))
	listen, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("%s\n", err)
		os.Exit(1)
	}
	defer listen.Close()

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			log.Printf("%s\n", err)
			os.Exit(1)
		}
		//处理请求
		if atomic.LoadInt32(&numberOfConnections) < int32(GConf.AgentConnectionLimit) {
			CreateConnection(conn)
			log.Printf("current there are %d agent connections\n", numberOfConnections)
		} else {
			//直接关闭连接
			conn.Close()
			log.Println("connections is too much now")
		}
	}
}
