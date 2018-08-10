package main

import (
	_ "sona/admin/routers"
	"github.com/astaxie/beego"
	"sona/common"
)

func main() {
	common.PrintLogo()
	beego.Run()
}

