#encoding=utf-8

import os
import mmap
import struct
import socket

#“产品线名.业务组名.服务名”组成serviceKey用于标识一个服务，各字段限长30字节
ServiceKeyCap = 92
#"section.key"组成服务的一个配置，各字段限长30字节
ConfKeyCap = 61
#value支持最大200字节
ConfValueCap = 200
ServiceConfLimit = 100
'''
一个bucket用于存放一个Service的配置，可存多少种Service
[版本号:serviceKey长度:serviceKey:配置个数:配置k1长度:配置k1:配置v1长度:配置v1:......]
[2:2:92:2:[2:61:2:30]:[2:61:2:30]:...]
'''
OneBucketCap = 2 + 2 + ServiceKeyCap + 2 + ServiceConfLimit * (2 + ConfKeyCap + 2 + ConfValueCap)
#一个service总共可包含100个配置
ServiceBucketLimit = 100
#配置总内存空间
#[Service1][Service2][Service3]...最多100个
TotalConfMemSize = OneBucketCap * ServiceBucketLimit

class SharedMemory(object):
    def __init__(self, service_key, index):
        self.fd = os.open('/tmp/sona/cfg.mmap', os.O_RDONLY)
        self.mmap = mmap.mmap(self.fd, TotalConfMemSize, flags = mmap.MAP_SHARED, prot = mmap.PROT_READ, offset = 0)
        self.service_key = service_key
        self.index = index
    '''
    if there are service configures
    '''
    def has_service(self):
        start = self.index * OneBucketCap
        self.mmap.seek(start)
        version_bytes = self.mmap.read(2)
        version = socket.ntohs(struct.unpack('H', version_bytes)[0])
        return version != 0
    '''
    get service key
    '''
    def get_service_key(self):
        start = self.index * OneBucketCap + 2
        self.mmap.seek(start)
        len_bytes = self.mmap.read(2)
        len = socket.ntohs(struct.unpack('H', len_bytes)[0])
        sk = self.mmap.read(len)
        return str(sk)
    '''
    get configure key
    '''
    def get_conf_key(self, pos):
        start = self.index * OneBucketCap + 6 + ServiceKeyCap + pos * (4 + ConfKeyCap + ConfValueCap)
        self.mmap.seek(start)
        len_bytes = self.mmap.read(2)
        len = socket.ntohs(struct.unpack('H', len_bytes)[0])
        conf_key = self.mmap.read(len)
        return str(conf_key)
    '''
    get configure value
    '''
    def get_conf_value(self, pos):
        start = self.index * OneBucketCap + 6 + ServiceKeyCap + pos * (4 + ConfKeyCap + ConfValueCap) + 2 + ConfKeyCap
        self.mmap.seek(start)
        len_bytes = self.mmap.read(2)
        len = socket.ntohs(struct.unpack('H', len_bytes)[0])
        conf_value = self.mmap.read(len)
        return str(conf_value)
    '''
    search a config key use binary search
    '''
    def search_one_conf(self, target_key):
        start = self.index * OneBucketCap + 4 + ServiceKeyCap
        self.mmap.seek(start)
        count_bytes = self.mmap.read(2)
        count = socket.ntohs(struct.unpack('H', count_bytes)[0])
        low, high = 0, count
        while low < high:
            mid = (low + high) / 2
            conf_key = self.get_conf_key(mid)
            if conf_key > target_key:
                high = mid
            elif conf_key < target_key:
                low = mid + 1
            else:
                return mid, True
        return -1, False
    '''
    get some key's value (public)
    '''
    def get_conf(self, conf_key):
        if not self.has_service() or self.get_service_key() != self.service_key:
            return ""
        pos, exist = self.search_one_conf(conf_key)
        if not exist:
            return ""
        return self.get_conf_value(pos)

'''
memory = SharedMemory('lebron.james.info', 0)
print(memory.get_conf("player.team"))
print(memory.get_conf("friends.list"))
memory = SharedMemory('miliao.milink.pushgateway', 1)
print(memory.get_conf("log.level"))
'''