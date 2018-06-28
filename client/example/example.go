package main

import (
	"easyconfig/client"
	"fmt"
	"time"
)

func main() {
	configClient, err := client.GetClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer configClient.Close()
	//方法1：订阅服务配置
	configClient.Subscribe("nba.player.info")
	//方法2：获取字符串
	value, err := configClient.Get("nba.player.info.lebron-james.number")
	if err == nil {
		fmt.Println("value is", value)
	}
	//方法3：获取列表
	list, err := configClient.GetList("nba.player.info.lebron-james.friends")
	for _, item := range list {
		fmt.Println(item)
	}
	time.Sleep(time.Second * 100)
}
