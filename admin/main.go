package main

import (
	_ "sona/admin/routers"
	"github.com/astaxie/beego"
	"fmt"
)

func main() {
	appname := beego.AppConfig.String("appname")
	fmt.Println(appname)
	beego.Run()
}

