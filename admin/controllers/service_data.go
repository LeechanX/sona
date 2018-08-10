package controllers

import (
    "strings"
    "sona/admin/models"
    "github.com/astaxie/beego"
)

type ServiceDataController struct {
    beego.Controller
}

func (c *ServiceDataController) Get() {
    serviceKey := c.Input().Get("servicekey")

    client, err := models.GetAdminClient()
    if err != nil {
        c.Data["Error"] = err.Error()
        c.TplName = "error.tpl"
        return
    }
    defer client.Close()
    serviceConf, err := models.Get(serviceKey, client)
    if err != nil {
        c.Data["Error"] = err.Error()
        c.TplName = "error.tpl"
        return
    }

    c.Data["ServiceKey"] = serviceConf.ServiceKey
    c.Data["Version"] = serviceConf.Version
    c.Data["Confs"] = serviceConf.Confs
    c.TplName = "info.tpl"
}

func (c *ServiceDataController) AddServiceKey() {
    serviceKey := c.GetString("product_name") + "." + c.GetString("group_name") + "." + c.GetString("service_name")
    client, err := models.GetAdminClient()
    if err != nil {
        c.Data["json"] = map[string]interface{}{"message": err.Error(), "code":300}
        c.ServeJSON()
        return
    }
    defer client.Close()
    //RPC
    err = models.Add(serviceKey, client)
    if err != nil {
        c.Data["json"] = map[string]interface{}{"message": err.Error(), "code":300}
        c.ServeJSON()
        return
    }
    c.Data["json"] = map[string]interface{}{"message": "Add successfully", "code":200}
    c.ServeJSON()
}

func (c *ServiceDataController) UpdateServiceKey() {
    serviceKey := c.GetString("servicekey")
    version, _ := c.GetUint32("version")
    data := c.GetStrings("confs")
    confs := make(map[string]string)
    for _, conf := range data {
        attrs := strings.Split(conf, " = ")
        key := strings.Trim(attrs[0], " ")
        value := strings.Trim(attrs[1], " ")
        confs[key] = value
    }
    client, err := models.GetAdminClient()
    if err != nil {
        c.Data["json"] = map[string]interface{}{"message": err.Error(), "code":300}
        c.ServeJSON()
        return
    }
    defer client.Close()
    //RPC
    err = models.Update(serviceKey, uint(version), confs, client)
    if err != nil {
        c.Data["json"] = map[string]interface{}{"message": err.Error(), "code":300}
    } else {
        c.Data["json"] = map[string]interface{}{"message": "Update successfully", "code":200}
    }
    c.ServeJSON()
}
