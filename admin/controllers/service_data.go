package controllers

import (
    "fmt"
    "github.com/astaxie/beego"
)

type ServiceDataController struct {
    beego.Controller
}

func (c *ServiceDataController) Get() {
    //TODO
    serviceKey := c.Input().Get("servicekey")
    c.Data["Service"] = serviceKey
    c.Data["Version"] = 3
    confs := make(map[string]string)
    confs["lebron.team"] = "lakers"
    confs["lebron.number"] = "23"
    confs["log.level"] = "debug"

    c.Data["Conf"] = confs
    c.TplName = "info.tpl"
}

func (c *ServiceDataController) AddServiceKey() {
    serviceKey := c.GetString("product_name") + "." + c.GetString("group_name") + "." + c.GetString("service_name")
    //RPC
    fmt.Println("add serviceKey ", serviceKey)
    c.Data["json"] = map[string]interface{}{"message": "hehehe", "code":200}
    c.ServeJSON()
}

func (c *ServiceDataController) UpdateServiceKey() {
    fmt.Println(c.GetString("serviceKey"))
    data := c.GetStrings("confs")
    for _, hehe := range data {
        fmt.Println("!!!!", hehe)
    }
    //RPC
    c.Data["json"] = map[string]interface{}{"message": "hehehe", "code":200}
    c.ServeJSON()
}
