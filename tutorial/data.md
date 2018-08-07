### 数据结构

配置按照服务名`service key`划分，一个服务下有若干个kv形式的配置项，并附带版本号

- `service key`：代表一个服务，格式为`product.group.service`，即`产品线.组.服务`
- 配置项：key是`section.configItem`，value是配置值，一个服务最多可有100个配置项
- 版本号：**对此服务配置的修改必须基于版本号进行CAS**

例：
服务`zhibo.pay.paynotify`（表示直播业务的支付小组的支付回调服务）下有配置，当前版本号为8：
```
log.level=debug
log.path=paylog/
store.domain=8.8.8.8
store.timeout=3000
hello.world=...
```

#### data in mongo DB
一个服务配置作为一个文档，如上示例：
```json
{"serviceKey":"zhibo.pay.paynotify",
"version":8,
"confkeys":["log.level","log.path","store.domain","store.timeout","hello.world"],
"confvalues":["debug","paylog/","8.8.8.8","3000","..."],
}
```
其中，字段`serviceKey`作为唯一索引
