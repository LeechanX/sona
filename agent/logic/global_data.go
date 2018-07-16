package logic

import (
    "sona/core"
    "sona/common/net/tcp/client"
    "sona/common/net/udp"
)

//记录每个service的最近心跳
var AccessRecordTable *AccessRecord

//配置管理
var ConfController *core.ConfigController

//broker客户端
var BrokerClient *client.Client

//biz服务
var BizServer *udp.Server