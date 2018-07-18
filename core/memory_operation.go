package core

import (
    "encoding/binary"
)

const (
    //“产品线名.业务组名.服务名”组成serviceKey用于标识一个服务，各字段限长30字节
    ServiceKeyCap = uint(92)
    //"section.key"组成服务的一个配置，各字段限长30字节
    ConfKeyCap = uint(61)
    ConfValueCap = uint(30)
    ServiceConfLimit = uint(100)

    //一个bucket用于存放一个Service的配置，可存多少种Service
    //[配置个数:配置k1长度:配置k1:配置v1长度:配置v1:......]
    //[2:[2:61:2:30]:[]:...]
    OneBucketCap = 2 + ServiceConfLimit * (2 + ConfKeyCap + 2 + ConfValueCap)

    //一个service总共可包含100个配置
    ServiceBucketLimit = uint(100)

    //配置总内存空间
    //[Service1][Service2][Service3]...最多100个
    TotalConfMemSize = OneBucketCap * ServiceBucketLimit

    //一个index信息
    //[serviceKey长度:serviceKey:版本号:位置]
    OneIndexCap = 2 + ServiceKeyCap + 2 + 2
    //索引总内存空间
    //前两个字节用于保存当前service个数
    TotalIndexMemSize = 2 + OneIndexCap * ServiceBucketLimit
)

//indexHub: [TotalMetaMemSize]byte

//获取当前service个数
func getServiceCount(indexHub *[TotalIndexMemSize]byte) uint {
    serviceCnt := binary.LittleEndian.Uint16(indexHub[:2])
    return uint(serviceCnt)
}

//获取对应位置的配置内容
func getServiceData(indexHub *[TotalIndexMemSize]byte, pos uint) (string, uint, uint) {
    start := 2 + OneIndexCap * pos
    keyLen := uint(binary.LittleEndian.Uint16(indexHub[start:start + 2]))
    serviceKey := string(indexHub[start + 2:start + 2 + keyLen])
    version := binary.LittleEndian.Uint16(indexHub[start + 2 + ServiceKeyCap:
        start + 2 + ServiceKeyCap + 2])
    index := binary.LittleEndian.Uint16(indexHub[start + 4 + ServiceKeyCap:
        start + 6 + ServiceKeyCap])
    return serviceKey, uint(version), uint(index)
}

//在indexHub上找到第一个 字典序>=serviceKey的service位置
//返回：位置，是否存在
func searchFirstNoLess(indexHub *[TotalIndexMemSize]byte, targetKey string) (uint, bool) {
    size := getServiceCount(indexHub)
    low, high := uint(0), size
    for low < high {
        mid := (low + high) / 2
        key, _, _ := getServiceData(indexHub, uint(mid))
        if key < targetKey {
            low = mid + 1
        } else if key > targetKey {
            if mid == 0 {
                return 0, false
            }
            prevKey, _, _ := getServiceData(indexHub, uint(mid - 1))
            if prevKey < targetKey {
                return mid, false
            } else {
                high = mid
            }
        } else {
            //找到了
            return mid, true
        }
    }
    return size, false
}

//设置pos位置处的service信息
func setService(indexHub *[TotalIndexMemSize]byte, serviceKey string, version uint, index uint, pos uint) {
    //[serviceKey长度:serviceKey:版本号:位置]
    //2 + ServiceKeyCap + 2 + 2
    start := 2 + OneIndexCap * uint(pos)
    //设置serviceKey长度
    binary.LittleEndian.PutUint16(indexHub[start:start + 2], uint16(len(serviceKey)))
    //设置serviceKey
    copy(indexHub[start + 2:start + 2 + ServiceKeyCap], serviceKey)
    //设置版本号
    binary.LittleEndian.PutUint16(indexHub[start + 2 + ServiceKeyCap:start + 4 + ServiceKeyCap], uint16(version))
    //设置位置
    binary.LittleEndian.PutUint16(indexHub[start + 4 + ServiceKeyCap:start + 6 + ServiceKeyCap], uint16(index))
}

//查找某service的信息，返回索引信息、版本
//-1表示不存在
func GetServiceIndex(indexHub *[TotalIndexMemSize]byte, serviceKey string) (int, uint) {
    pos, exist := searchFirstNoLess(indexHub, serviceKey)
    if !exist {
        return -1, 0
    }
    start := 2 + OneIndexCap * uint(pos) + 2 + ServiceKeyCap
    version := uint(binary.LittleEndian.Uint16(indexHub[start:start + 2]))
    index := int(binary.LittleEndian.Uint16(indexHub[start + 2:start + 2 + 2]))
    return index, version
}

//插入新service
func InsertService(indexHub *[TotalIndexMemSize]byte, serviceKey string, version uint, index uint) bool {
    count := getServiceCount(indexHub)
    pos, exist := searchFirstNoLess(indexHub, serviceKey)
    if exist {
        //已存在，改个版本号就行了
        start := 2 + OneIndexCap * uint(pos) + 2 + ServiceKeyCap
        binary.LittleEndian.PutUint16(indexHub[start:start + 2], uint16(version))
        return true
    }
    if count == ServiceBucketLimit {
        //已经满了
        return false
    }
    if pos == count {
        //添加到indexHub尾端
        setService(indexHub, serviceKey, version, index, count)
    } else {
        start := 2 + OneIndexCap * uint(pos)
        //pos位置处整体后移
        copy(indexHub[start + OneIndexCap:start + OneIndexCap * (count + 1 - pos)],
            indexHub[start:start + OneIndexCap * (count - pos)])
        //添加到pos位置
        setService(indexHub, serviceKey, version, index, pos)
    }
    //service Count加1
    binary.LittleEndian.PutUint16(indexHub[:2], uint16(count + 1))
    return true
}

//删除service
func RemoveService(indexHub *[TotalIndexMemSize]byte, serviceKey string) {
    pos, exist := searchFirstNoLess(indexHub, serviceKey)
    if !exist {
        return
    }
    count := getServiceCount(indexHub)
    if pos != count - 1 {
        start := 2 + OneIndexCap * pos
        //不是最后一个，后面的整体向前移动
        behind := count - 1 - pos
        copy(indexHub[start:start + OneIndexCap * behind],
            indexHub[start + OneIndexCap:start + OneIndexCap + OneIndexCap * behind])
    }
    //service Count减1
    binary.LittleEndian.PutUint16(indexHub[:2], uint16(count - 1))
}

//更新某service版本
func UpdServiceVersion(indexHub *[TotalIndexMemSize]byte, serviceKey string, version uint) {
    pos, exist := searchFirstNoLess(indexHub, serviceKey)
    if exist {
        start := 2 + OneIndexCap * uint(pos) + 2 + ServiceKeyCap
        binary.LittleEndian.PutUint16(indexHub[start:start + 2], uint16(version))
    }
}

//获取所有service和对应版本
func GetAllService(indexHub *[TotalIndexMemSize]byte) map[string]uint {
    count := getServiceCount(indexHub)
    services := make(map[string]uint)
    for i := uint(0);i < count;i++ {
        serviceKey, version, _ := getServiceData(indexHub, i)
        services[serviceKey] = version
    }
    return services
}

//获取第一个可用index
func GetFirstIndexFree(indexHub *[TotalIndexMemSize]byte) uint {
    count := getServiceCount(indexHub)
    idxSet := make(map[uint]bool)
    for i := uint(0);i < count;i++ {
        _, _, idx := getServiceData(indexHub, i)
        idxSet[idx] = true
    }
    for idx := uint(0);idx < ServiceBucketLimit;idx++ {
        if _, ok := idxSet[idx];!ok {
            return idx
        }
    }
    return ServiceBucketLimit
}

//一个bucket用于存放一个Service的配置，可存多少种Service
//[配置个数:配置k1长度:配置k1:配置v1长度:配置v1:......]
//[2:[1:61:1:30]:[]:...]
//OneBucketCap = + 2 + ServiceConfLimit * (1 + ConfKeyCap + 1 + ConfValueCap)
//confHub: [TotalConfMemSize]byte

//获取某service配置个数
func GetConfCount(confHub *[TotalConfMemSize]byte, idx uint) uint {
    start := idx * OneBucketCap
    confCnt := uint(binary.LittleEndian.Uint16(confHub[start:start + 2]))
    return confCnt
}

//获取某service的某位置pos上的配置
func getOneConf(confHub *[TotalConfMemSize]byte, idx uint, pos uint) (string, string) {
    start := idx * OneBucketCap + 2 + pos * (2 + ConfKeyCap + 2 + ConfValueCap)
    keyLen := uint(binary.LittleEndian.Uint16(confHub[start:start + 2]))
    confKey := string(confHub[start + 2:start + 2 + keyLen])
    valueLen := uint(binary.LittleEndian.Uint16(confHub[start + 2 + ConfKeyCap:start + 2 + ConfKeyCap + 2]))
    confValue := string(confHub[start + 2 + ConfKeyCap + 2:start + 2 + ConfKeyCap + 2 + valueLen])
    return confKey, confValue
}

//为某service在pos位置处设置配置
func setOneConf(confHub *[TotalConfMemSize]byte, idx uint, pos uint, confKey string, value string) {
    var start = idx * OneBucketCap + 2 + pos * (2 + ConfKeyCap + 2 + ConfValueCap)
    binary.LittleEndian.PutUint16(confHub[start:start + 2], uint16(len(confKey)))
    start += 2
    copy(confHub[start:start + ConfKeyCap], confKey)
    start += ConfKeyCap
    binary.LittleEndian.PutUint16(confHub[start:start + 2], uint16(len(value)))
    start += 2
    copy(confHub[start:start + ConfValueCap], value)
}

//获取某service的一个配置值
func GetConf(confHub *[TotalConfMemSize]byte, idx uint, confKey string) string {
    pos, exist := searchOneConf(confHub, idx, confKey)
    if !exist {
        return ""
    }
    key, value := getOneConf(confHub, idx, pos)
    if key != confKey {
        return ""
    }
    return value
}

//在某service下的配置中，找到第一个 字典序>=confKey的成员
//返回位置
func searchOneConf(confHub *[TotalConfMemSize]byte, idx uint, confKey string) (uint, bool) {
    confCnt := GetConfCount(confHub, idx)
    low, high := uint(0), confCnt
    for low < high {
        mid := (low + high) / 2
        key, _ := getOneConf(confHub, idx, mid)
        if key < confKey {
            low = mid + 1
        } else if key > confKey {
            if mid == 0 {
                return 0, false
            }
            prevKey, _ := getOneConf(confHub, idx, mid - 1)
            if prevKey >= confKey {
                high = mid
            } else {
                return mid, false
            }
        } else {
            return mid, true
        }
    }
    return confCnt, false
}

//获取某service的全部配置
func GetServiceConf(confHub *[TotalConfMemSize]byte, idx uint) map[string]string {
    confCnt := GetConfCount(confHub, idx)
    allConf := make(map[string]string)
    for i := uint(0);i < confCnt;i++ {
        k, v := getOneConf(confHub, idx, i)
        allConf[k] = v
    }
    return allConf
}

//添加一个service的配置，前提：confKey已按照字典序排序
func AddServiceConf(confHub *[TotalConfMemSize]byte, idx uint, confKeys []string, values []string) {
    var start = idx * OneBucketCap
    confCnt := uint(len(confKeys))
    //设置长度
    binary.LittleEndian.PutUint16(confHub[start:start + 2], uint16(confCnt))
    start += 2
    for i := uint(0);i < confCnt;i++ {
        setOneConf(confHub, idx, i, confKeys[i], values[i])
    }
}

//为某service添加、修改一个配置
func AddOrUpdateConf(confHub *[TotalConfMemSize]byte, idx uint, confKey string, value string) bool {
    confCnt := GetConfCount(confHub, idx)
    pos, exist := searchOneConf(confHub, idx, confKey)
    if exist {
        //修改：重设置值
        setOneConf(confHub, idx, pos, confKey, value)
        return true
    }
    //新增
    if confCnt == ServiceConfLimit {
        //已经满了
        return false
    }
    if pos == confCnt {
        //尾部添加即可
        setOneConf(confHub, idx, confCnt, confKey, value)
    } else {
        //整体后移
        start := idx * OneBucketCap + pos * (2 + ConfKeyCap + 2 + ConfValueCap)
        copy(confHub[start + (2 + ConfKeyCap + 2 + ConfValueCap):
            start + (2 + ConfKeyCap + 2 + ConfValueCap) * (confCnt - pos + 1)],
            confHub[start:
                start + (2 + ConfKeyCap + 2 + ConfValueCap) * (confCnt - pos)])
        setOneConf(confHub, idx, pos, confKey, value)
    }
    //配置个数+1
    start := idx * OneBucketCap
    binary.LittleEndian.PutUint16(confHub[start:start + 2], uint16(confCnt + 1))
    return true
}

//删除某service的配置
func RemoveServiceConf(confHub *[TotalConfMemSize]byte, idx uint) {
    start := idx * OneBucketCap
    binary.LittleEndian.PutUint16(confHub[start:start + 2], 0)
}

//为某service删除一个配置
func RemoveOneConf(confHub *[TotalConfMemSize]byte, idx uint, confKey string) {
    confCnt := GetConfCount(confHub, idx)
    pos, exist := searchOneConf(confHub, idx, confKey)
    if !exist {
        return
    }
    if pos != confCnt - 1 {
        //整体前移
        start := idx * OneBucketCap + pos * (2 + ConfKeyCap + 2 + ConfValueCap)
        behind := confCnt - 1 - pos
        copy(confHub[start:start + (2 + ConfKeyCap + 2 + ConfValueCap) * behind],
            confHub[start + (2 + ConfKeyCap + 2 + ConfValueCap):
                start + (2 + ConfKeyCap + 2 + ConfValueCap) * (behind - 1)])
    }
    //配置个数-1
    start := idx * OneBucketCap
    binary.LittleEndian.PutUint16(confHub[start:start + 2], uint16(confCnt - 1))
}
