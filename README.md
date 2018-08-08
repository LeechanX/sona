## Sona

Sona是一个go语言实现的高效、实时、高可用的linux分布式配置中心，轻量级支持（Golang/C++/Java/Python）主流编程语言接口

```
    ___  ___  _ __   __ _ 
   / __|/ _ \| '_ \ / _` |
   \__ \ (_) | | | | (_| |   
   |___/\___/|_| |_|\__,_|   

```

## 特点

sona配置中心采用了经典一中心(broker)多agent的分布式架构，基于共享内存下发、存储各节点所需配置，为业务提供KV方式访问（最新）配置

- `高度可用`：agent即使挂掉也不影响已有业务读配置，而broker以keepalived组件保证其高可用
- `一致性`：broker采用主备模式，仅主对外服务，正常情况下保证数据完全一致；
仅在主备切换时可能有短时间最新数据的延迟。总体而言实现了数据的最终一致性
- `实时更新`：正常情况下，数据实时更新到各节点；仅在主备切换时刻，可能有短时间的数据延迟
- `API简单`：无配置文件概念，故业务无需关心配置文件解析；
数据实时更新对业务完全透明，业务无需编写配置更新的回调函数


## USAGE

提供C/C++/Java/Python/Golang多语言支持，基于protobuf-2.6.1

**go语言：**

```
import "sona/api"

configApi, err := api.GetApi("nba.player.info") //获取nba.player.info服务的配置
if err == nil {
    defer configApi.Close()

    value := configApi.Get("lebron-james","number") //获取lebron-james.number值 (string)

    list := configApi.GetList("lebron-james","friends") //获取lebron-james.friends值列表 ([]string)
}
```
**C++：** 见目录api/cpp

```
#include "sona_api.h"

sona_api* api = init_api("nba.player.info"); //获取nba.player.info服务的配置
if api != NULL {
    string value = api->get("lebron-james", "number"); //获取lebron-james.number值 (string)
    vector<string> list = api->get_list("lebron-james","friends"); //获取lebron-james.friends值列表 (vector<string>)
}
```
**Java：** jar包在api/java/sona_api/lib/sona_api.jar，源码见api/java/sona_api/src
```
import org.sona.api.SonaApi;

SonaApi api = null;
try {
    api = new SonaApi("lebron.james.info"); //获取lebron.james.info服务的配置
} catch (Exception e) {
}

if (api != null) {
    String value = api.Get("player", "number"); //获取player.number值 (string)
    System.out.println(value);
    ArrayList<String> list = api.GetList("friends", "list"); //获取friends.list值列表 (ArrayList<string>)
    for (String item: list) {
        System.out.println(item);
    }
}
```

**Python：** python2.7、python3.6兼容的代码（如果要使用python3，请重新编译protocol/base_protocol.proto并将产生的base_protocol_pb2.py替换到api/python/sona目录下）
```
from sona import api

try:
    api = api.SonaApi("lebron.james.info")
except Exception as e:
    api = None
    print(e)
    
if api:
    print(api.get("player", "team")) #获取player.number值 (string)
    print(api.get_list("friends", "list")) #获取friends.list值列表 (List<string>)
```


## 原理介绍

![arch](tutorial/pictures/arch.jpg)

- `broker`做数据中心管控，管理配置数据的增、改、订阅、下发，使用keepalived做高可用
- `agent`做各节点配置代理，管理下发到各个节点的配置、业务订阅配置等
- 共享内存是`sona`核心结构，`agent`实际在这里管理配置，业务实际也从这里读取最新配置

数据介绍 [data readme][1]

[1]: https://github.com/LeechanX/Sona/blob/master/tutorial/data.md

共享内存模型与agent、api [mem readme][2]

[2]: https://github.com/LeechanX/Sona/blob/master/tutorial/mem.md

broker介绍 [broker readme][3]

[3]: https://github.com/LeechanX/Sona/blob/master/tutorial/broker.md

## 安装

依赖：

1、mongoDB（可单点，但建议主备）

2、keepalived（可不安装，如果不在意broker单点的话）

安装流程：

- 将本项目git clone到您的`$GOPATH`下
- 进入`$GOPATH/sona`目录下
- 执行`make`进行编译，所有可执行文件sona_agent、sona_broker将产生在`bin`目录下
- 编辑broker配置文件（示例在`broker/conf/`目录下），后执行`bin/sona_broker -c 配置文件路径`启动broker
- 编辑agent配置文件（示例在`agent/conf/`目录下），后执行`bin/sona_agent -c 配置文件路径`启动agent
- 所有语言的api位于`api`目录下，自行使用即可

