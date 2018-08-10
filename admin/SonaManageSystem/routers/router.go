package routers

import (
	"sona/admin/SonaManageSystem/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
    beego.Router("/add", &controllers.MainController{}, "*:Add")

    beego.Router("/run/query", &controllers.ServiceDataController{})
    beego.Router("/run/add", &controllers.ServiceDataController{}, "post:AddServiceKey")
    beego.Router("/run/update", &controllers.ServiceDataController{}, "post:UpdateServiceKey")
}
