package main

import (
    "fmt"
    "time"
    "sona/client"
)

func main() {
    //获取service = nba.player.info的服务配置
    configClient, err := client.GetClient("nba.player.info")
    if err != nil {
        fmt.Println(err)
        return
    }
    defer configClient.Close()
    //获取值
    value, err := configClient.Get("lebron-james","number")
    if err == nil {
        fmt.Println("value is", value)
    }
    //获取列表
    list, err := configClient.GetList("lebron-james","friends")
    for _, item := range list {
        fmt.Println(item)
    }
    time.Sleep(time.Second * 100)
}
