package logic

import (
    "log"
    "time"
    "sona/broker/dao"
)

func ReloadData() {
    for {
        //一分钟重载一次全部配置
        time.Sleep(time.Second * 60)
        data, err := dao.ReloadAllData()
        if err != nil {
            log.Printf("when reload data: %s\n", err)
            continue
        }
        //TODO 重设全局数据
    }
}