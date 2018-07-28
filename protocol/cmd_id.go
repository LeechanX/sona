package protocol

const (
    KeepUsingReqId         = uint(1)//client向agent发起使用中请求
    SubscribeReqId         = 2//client向agent发起订阅请求,agent向broker发起订阅请求
    SubscribeBrokerRspId   = 3//broker向agent回复订阅请求
    SubscribeAgentRspId    = 4//agent向client回复订阅请求
    PushServiceConfigReqId = 5//broker推送给agent某服务的最新配置
    PullServiceConfigReqId = 6//agent向broker拉取一个服务的最新配置
    PullServiceConfigRspId = 7//broker回复agent一个服务的最新配置

    AdminAddConfigReqId   = 21//admin向broker发起新增命令
    AdminCleanConfigReqId = 22//admin向broker发起删除命令
    AdminUpdConfigReqId   = 23//admin向broker发起修改命令
    AdminExecuteRspId     = 24//broker回复admin执行结果
    AdminGetConfigReqId   = 25//admin向broker发起获取请求
    AdminGetConfigRspId   = 26//admin向broker发起获取请求
)