package main

import (
    "os"
    "fmt"
    "flag"
    "time"
    "bufio"
    "strings"
    "sona/core"
    "sona/protocol"
    "sona/common/net/tcp/client"
    "github.com/golang/protobuf/proto"
)

type ServiceConfig struct {
    serviceKey string
    version uint
    conf map[string]string
}

func parseCommand(line string) []string {
    result := make([]string, 0)
    line = strings.Trim(line, " ")
    lineAttrs := strings.Split(line, " ")
    for _, attr := range lineAttrs {
        result = append(result, strings.Trim(attr, " "))
    }
    return result
}

func get(serviceKey string, client *client.SyncClient) *ServiceConfig {
    req := &protocol.AdminGetConfigReq{}
    req.ServiceKey = proto.String(serviceKey)

    err := client.Send(protocol.AdminGetConfigReqId, req)
    if err != nil {
        fmt.Println(err)
        return nil
    }
    //接收 100ms超时
    timeout := 100 * time.Millisecond
    rsp := &protocol.AdminGetConfigRsp{}
    err = client.Read(timeout, protocol.AdminGetConfigRspId, rsp)
    if err != nil {
        fmt.Println(err)
        return nil
    }
    if *rsp.Code == -1 {
        fmt.Println("this service is not exist")
        return nil
    }

    fmt.Println()
    serviceConf := &ServiceConfig{
        serviceKey:*rsp.ServiceKey,
        version:uint(*rsp.Version),
        conf:make(map[string]string),
    }
    for i := 0;i < len(rsp.ConfKeys);i++ {
        key, value := rsp.ConfKeys[i], rsp.Values[i]
        serviceConf.conf[key] = value
    }

    fmt.Printf("%-60s", "service key")
    fmt.Printf(" %s\n", serviceConf.serviceKey)
    fmt.Printf("%-60s", "version")
    fmt.Printf(" %d\n", serviceConf.version)
    fmt.Println()
    for confKey, confValue := range serviceConf.conf {
        fmt.Printf("%-60s", confKey)
        fmt.Printf(" %s\n", confValue)
    }
    fmt.Println()
    return serviceConf
}

func add(serviceKey string, client *client.SyncClient) {
    confKeys := make([]string, 0)
    confValues := make([]string, 0)
    for {
        var confKey string
        fmt.Printf("input a configure key (q to exit): ")
        fmt.Scanf("%s", &confKey)
        if confKey == "q" {
            break
        }
        if !core.IsValidityConfKey(confKey) {
            fmt.Println("key format error")
            continue
        }
        //check format
        fmt.Printf("input the configure value: ")
        var confValue string
        fmt.Scanf("%s", &confValue)

        confKeys = append(confKeys, confKey)
        confValues = append(confValues, confValue)
    }

    //send add request
    req := &protocol.AdminAddConfigReq{}
    req.ServiceKey = proto.String(serviceKey)
    req.ConfKeys = confKeys
    req.Values = confValues

    err := client.Send(protocol.AdminAddConfigReqId, req)
    if err != nil {
        fmt.Println(err)
        return
    }

    //接收 100ms超时
    timeout := 100 * time.Millisecond
    rsp := &protocol.AdminExecuteRsp{}
    err = client.Read(timeout, protocol.AdminExecuteRspId, rsp)
    if err != nil {
        fmt.Println(err)
        return
    }
    if *rsp.Code == 0 {
        fmt.Println("Add successfully")
    } else {
        fmt.Println(err)
    }
}

func del(serviceKey string, client *client.SyncClient) {
    serviceConf := get(serviceKey, client)
    if serviceConf == nil {
        return
    }
    fmt.Printf("Delete? (y/n): ")
    var ensure string
    fmt.Scanf("%s", &ensure)
    if ensure == "y" || ensure == "Y" {
        req := &protocol.AdminDelConfigReq{}
        req.ServiceKey = proto.String(serviceKey)
        req.Version = proto.Uint32(uint32(serviceConf.version))
        err := client.Send(protocol.AdminDelConfigReqId, req)
        if err != nil {
            fmt.Println(err)
            return
        }
        //接收 100ms超时
        timeout := 100 * time.Millisecond
        rsp := &protocol.AdminExecuteRsp{}
        err = client.Read(timeout, protocol.AdminExecuteRspId, rsp)
        if err != nil {
            fmt.Println(err)
            return
        }
        if *rsp.Code == 0 {
            fmt.Println("Delete successfully")
        } else {
            fmt.Println(err)
        }
    }
}

func update(serviceKey string, client *client.SyncClient) {
    serviceConf := get(serviceKey, client)
    if serviceConf == nil {
        return
    }

    fmt.Println(">> command mode ")
    fmt.Println(">>  ")
    fmt.Println(">> add key value: add or update key value to this service")
    fmt.Println(">> del key: delete key from this service")
    fmt.Println(">> quit: finish and leave command mode")
    fmt.Println(">> ")

    reader := bufio.NewReader(os.Stdin)
    var line string
    for {
        fmt.Printf(">> ")
        line, _ = reader.ReadString('\n')
        line = strings.TrimSuffix(line, "\n")
        command := parseCommand(line)
        if len(command) == 0 {
            continue
        }

        if command[0] == "quit" {
            break
        }

        if command[0] == "add" {
            if len(command) != 3 {
                fmt.Println(">> add format error")
            } else {
                key, value := command[1], command[2]
                serviceConf.conf[key] = value
                fmt.Println(">> ok")
            }
        } else if command[0] == "del" {
            if len(command) != 2 {
                fmt.Println(">> del format error")
            } else {
                key := command[1]
                if _, ok := serviceConf.conf[key];ok {
                    delete(serviceConf.conf, key)
                    fmt.Println(">> ok")
                } else {
                    fmt.Println(">> no this key")
                }
            }
        }
    }

    fmt.Println()
    //build PB
    req := &protocol.AdminUpdConfigReq{}
    req.ServiceKey = proto.String(serviceConf.serviceKey)
    req.Version = proto.Uint32(uint32(serviceConf.version))
    req.ConfKeys = make([]string, 0)
    req.Values = make([]string, 0)

    for confKey, confValue := range serviceConf.conf {
        fmt.Printf("%-60s", confKey)
        fmt.Printf(" %s\n", confValue)
        req.ConfKeys = append(req.ConfKeys, confKey)
        req.Values = append(req.Values, confValue)
    }
    fmt.Println()

    fmt.Printf("Submit? (y/n): ")
    var ensure string
    fmt.Scanf("%s", &ensure)
    if ensure == "y" || ensure == "Y" {
        err := client.Send(protocol.AdminUpdConfigReqId, req)
        if err != nil {
            fmt.Println(err)
            return
        }
        //接收 100ms超时
        timeout := 100 * time.Millisecond
        rsp := &protocol.AdminExecuteRsp{}
        err = client.Read(timeout, protocol.AdminExecuteRspId, rsp)
        if err != nil {
            fmt.Println(err)
            return
        }
        if *rsp.Code == 0 {
            fmt.Println("Submit successfully")
        } else {
            fmt.Println(err)
        }
    }
}

func main() {
    host := flag.String("host", "", "admin server ip")
    port := flag.Int("port", 0, "admin server port")
    operation := flag.String("operation", "", "[get],[add] or [update],[delete] configures")

    flag.Parse()

    if *host == "" || *port == 0 {
        fmt.Println("no host or port is specified")
        return
    }
    if *operation != "get" && *operation != "add" && *operation != "update" && *operation != "delete" {
        fmt.Println("only support operations: [get],[add] or [update],[delete]")
        return
    }

    //connect to admin server
    cli, err := client.CreateSyncClient(*host, *port)
    if err != nil {
        fmt.Println(err)
        return
    }

    //ui
    var serviceKey string
    for {
        fmt.Printf("input service key: ")
        fmt.Scanf("%s", &serviceKey)
        //check format
        if !core.IsValidityServiceKey(serviceKey) {
            fmt.Println("service key format error")
        } else {
            break
        }
    }

    if *operation == "get" {
        get(serviceKey, cli)
    } else if *operation == "add" {
        add(serviceKey, cli)
    } else if *operation == "update" {
        update(serviceKey, cli)
    } else {
        del(serviceKey, cli)
    }
}
