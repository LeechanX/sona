#### client端行为

在config hub里获取，没有就给agent发拉取消息，然后在hub上轮询100ms：一直没有配置agent就很繁忙+业务总阻塞
超时返回不存在

#### agent端行为

（1）接收client拉取消息，交给写线程 W goroutine
（2）接收broker消息，并更新 R goroutine
（3）周期性更新bucket T goroutine

 R1 (UDP) -> W
 R2 (TCP)
 T -> W
 T独立

 agent问题：TCP粘包分包、错误处理、重连