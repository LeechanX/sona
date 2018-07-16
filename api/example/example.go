package main

import (
    "fmt"
    "sona/api"
)

func main() {
    //获取service = nba.player.info的服务配置
    configApi, err := api.GetApi("nba.player.info")
    if err != nil {
        fmt.Println(err)
        return
    }
    defer configApi.Close()
    //获取值
    value := configApi.Get("lebron-james","number")
    fmt.Println("value is", value)

    //获取列表
    list := configApi.GetList("lebron-james","friends")
    for _, item := range list {
        fmt.Println(item)
    }
}
