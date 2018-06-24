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
	value, err := configClient.Get("lebron.jamses.number")
	if err == nil {
		fmt.Println("value is", value)
	}
	time.Sleep(time.Second * 100)
}
