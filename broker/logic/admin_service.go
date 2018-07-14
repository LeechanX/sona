package logic

import (
    "os"
    "fmt"
    "log"
    "net"
    "sync/atomic"
    "sona/broker/conf"
)

//新增配置
func AddConfig(serviceKey string, configKeys []string, values []string) error {
    newVersion, err := ConfigData.AddConfig(serviceKey, configKeys, values)
    if err != nil {
        log.Printf("add %s configure meet error: %s\n", serviceKey, err)
        return err
    }
    //如果有agent订阅, push给每个agent连接
    agents := SubscribedBook.GetSubscribers(serviceKey)
    for _, agent := range agents {
        agent.PushConfig(serviceKey, uint32(newVersion), configKeys, values)
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
    agents := SubscribedBook.GetSubscribers(serviceKey)
    for _, agent := range agents {
        agent.PushConfig(serviceKey, uint32(newVersion), configKeys, values)
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
    agents := SubscribedBook.GetSubscribers(serviceKey)
    for _, agent := range agents {
        agent.PushConfig(serviceKey, uint32(newVersion), []string{}, []string{})
    }
    return nil
}

func AdminService() {
    tcpAddr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf("0.0.0.0:%d", conf.GlobalConf.AdminPort))
    listen, err := net.ListenTCP("tcp", tcpAddr)
    if err != nil {
        log.Fatalf("%s\n", err)
        os.Exit(1)
    }
    log.Printf("create admin server(%s) successfully\n", fmt.Sprintf("0.0.0.0:%d", conf.GlobalConf.AdminPort))
    defer listen.Close()

    for {
        conn, err := listen.AcceptTCP()
        if err != nil {
            log.Printf("%s\n", err)
            os.Exit(1)
        }
        //处理请求
        if atomic.LoadInt32(&numberOfConnections) < int32(conf.GlobalConf.AgentConnectionLimit) {
            //TODO
            //(conn)
            log.Printf("current there are %d agent connections\n", numberOfConnections)
        } else {
            //直接关闭连接
            conn.Close()
            log.Println("connections is too much now")
        }
    }
}
