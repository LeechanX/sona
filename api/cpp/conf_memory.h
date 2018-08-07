#ifndef __CONF_MEMORY_H__
#define __CONF_MEMORY_H__

//“产品线名.业务组名.服务名”组成serviceKey用于标识一个服务，各字段限长30字节
#define ServiceKeyCap 92
//"section.key"组成服务的一个配置，各字段限长30字节
#define ConfKeyCap 61
//value支持最大200字节
#define ConfValueCap 200
#define ServiceConfLimit 100
//一个bucket用于存放一个Service的配置，可存多少种Service
//[版本号:serviceKey长度:serviceKey:配置个数:配置k1长度:配置k1:配置v1长度:配置v1:......]
//[2:2:92:2:[2:61:2:30]:[2:61:2:30]:...]
#define OneBucketCap (6 + ServiceKeyCap + ServiceConfLimit * (4 + ConfKeyCap + ConfValueCap))
//一个service总共可包含100个配置
#define ServiceBucketLimit 100
//配置总内存空间 [Service1][Service2][Service3]...最多100个
#define TotalConfMemSize (OneBucketCap * ServiceBucketLimit)

int get_conf(const char* memory, unsigned index,
        const char* service_key, const char* key, char* value);
void* attach_mmap();
void detach_mmap(void* m);

#endif
