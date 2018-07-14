## Sona

Sona是是一个go语言实现的高效、实时、高可用的分布式配置中心，支持（C/C++/Go/Java/Python）接口主流编程语言

```
    ___  ___  _ __   __ _ 
   / __|/ _ \| '_ \ / _` |
   \__ \ (_) | | | | (_| |   
   |___/\___/|_| |_|\__,_|   


```

author by: leechanx<br/>
contact: leechan8@outlook.com<br/>
wechat: leechanx<br/>

## 特点

#### 1. 以配置项（key=value）为单位进行配置管理：
- 抛弃书写复杂且容易写错导致解析错误的配置文件方式管理
#### 2. 基于共享内存管理各节点的配置
- 在共享内存上设计通用高效的数据结构，轻松应对高并发访问，方便地支持多编程语言C/C++/Go/Java/Python接口
- 即使配置服务挂掉，也几乎不影响业务服务本身
- 获取一个配置项的最新值和几乎和访问变量一样毫无成本，可以实时得到配置最新值

## 架构

TODO

## 安装

TODO
