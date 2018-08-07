package core

import "encoding/binary"

const (
    //“产品线名.业务组名.服务名”组成serviceKey用于标识一个服务，各字段限长30字节
    ServiceKeyCap = uint(92)
    //"section.key"组成服务的一个配置，各字段限长30字节
    ConfKeyCap = uint(61)
    //value支持最大200字节
    ConfValueCap = uint(200)
    ServiceConfLimit = uint(100)
    //一个bucket用于存放一个Service的配置，可存多少种Service
    //[版本号:serviceKey长度:serviceKey:配置个数:配置k1长度:配置k1:配置v1长度:配置v1:......]
    //[2:2:92:2:[2:61:2:30]:[2:61:2:30]:...]
    OneBucketCap = 2 + 2 + ServiceKeyCap + 2 + ServiceConfLimit * (2 + ConfKeyCap + 2 + ConfValueCap)
    //一个service总共可包含100个配置
    ServiceBucketLimit = uint(100)
    //配置总内存空间
    //[Service1][Service2][Service3]...最多100个
    TotalConfMemSize = OneBucketCap * ServiceBucketLimit
)

//在指定切片（2字节）上获取数字值
func getNumber(slice []byte) uint {
    number := binary.BigEndian.Uint16(slice)
    return uint(number)
}

//在指定切片（2字节）上写入数字值
func setNumber(slice []byte, v uint16) {
    binary.BigEndian.PutUint16(slice, v)
}

//获取目前内存中所有service配置所在索引位置
func GetAllServiceIndex(memory *[TotalConfMemSize]byte) map[string]uint {
    start := uint(0)
    serviceIndex := make(map[string]uint)
    var index uint = 0
    for start < TotalConfMemSize {
        if getNumber(memory[start:start + 2]) > 0 {
            //说明有配置
            //获取serviceKey长度
            serviceKeyLen := getNumber(memory[start + 2:start + 4])
            //获取serviceKey
            serviceKey := string(memory[start + 4:start + 4 + serviceKeyLen])
            //fmt.Printf("get %s\n", serviceKey)
            serviceIndex[serviceKey] = index
        }
        start += OneBucketCap
        index += 1
    }
    return serviceIndex
}

//接下来是某service具体信息的操作
//获取某service里的配置个数
func GetConfCount(memory *[TotalConfMemSize]byte, idx uint) uint {
    start := idx * OneBucketCap + 4 + ServiceKeyCap
    return getNumber(memory[start:start + 2])
}

//为某service在pos位置处设置配置
func setOneConf(memory *[TotalConfMemSize]byte, idx uint, pos uint, confKey string, value string) {
    var start = idx * OneBucketCap + 2 + 2 + ServiceKeyCap + 2 + pos * (2 + ConfKeyCap + 2 + ConfValueCap)
    setNumber(memory[start:start + 2], uint16(len(confKey)))
    start += 2
    copy(memory[start:start + ConfKeyCap], confKey)
    start += ConfKeyCap
    setNumber(memory[start:start + 2], uint16(len(value)))
    start += 2
    copy(memory[start:start + ConfValueCap], value)
}

//获取某service的某位置pos上的配置
func getOneConf(memory *[TotalConfMemSize]byte, idx uint, pos uint) (string, string) {
    start := idx * OneBucketCap + 2 + 2 + ServiceKeyCap + 2 + pos * (2 + ConfKeyCap + 2 + ConfValueCap)
    keyLen := getNumber(memory[start:start + 2])
    confKey := string(memory[start + 2:start + 2 + keyLen])
    valueLen := getNumber(memory[start + 2 + ConfKeyCap:start + 2 + ConfKeyCap + 2])
    confValue := string(memory[start + 2 + ConfKeyCap + 2:start + 2 + ConfKeyCap + 2 + valueLen])
    return confKey, confValue
}

//某位置是否有配置
func HasService(memory *[TotalConfMemSize]byte, idx uint) bool {
    start := idx * OneBucketCap
    return getNumber(memory[start:start + 2]) != 0
}

//获取某位置上的serviceKey与版本
func GetServiceKey(memory *[TotalConfMemSize]byte, idx uint) (string, uint) {
    if !HasService(memory, idx) {
        return "", 0
    }
    start := idx * OneBucketCap
    version := getNumber(memory[start:start + 2])
    serviceKeyLen := getNumber(memory[start + 2:start + 4])
    //获取serviceKey
    serviceKey := string(memory[start + 4:start + 4 + serviceKeyLen])
    return serviceKey, version
}

//获取某service的某配置confKey的值
func GetConf(memory *[TotalConfMemSize]byte, serviceKey string, idx uint, confKey string) string {
    if !HasService(memory, idx) {
        return ""
    }
    //检查：idx位置处是否是要求的serviceKey
    serviceKeyInMem, _ := GetServiceKey(memory, idx)
    if serviceKeyInMem != serviceKey {
        return ""
    }
    pos, exist := searchOneConf(memory, idx, confKey)
    if !exist {
        return ""
    }
    key, value := getOneConf(memory, idx, pos)
    if key != confKey {
        return ""
    }
    return value
}

//在某service下的配置中，找到第一个 字典序>=confKey的成员
//返回位置
func searchOneConf(memory *[TotalConfMemSize]byte, idx uint, confKey string) (uint, bool) {
    confCnt := GetConfCount(memory, idx)
    low, high := uint(0), confCnt
    for low < high {
        mid := (low + high) / 2
        key, _ := getOneConf(memory, idx, mid)
        if key < confKey {
            low = mid + 1
        } else if key > confKey {
            if mid == 0 {
                return 0, false
            }
            prevKey, _ := getOneConf(memory, idx, mid - 1)
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

//添加一个service的配置，前提：confKey已按照字典序排序
func AddServiceConf(memory *[TotalConfMemSize]byte, idx uint,
    serviceKey string, version uint, confKeys []string, values []string) {
    var start = idx * OneBucketCap
    confCnt := uint(len(confKeys))
    serviceKeyLen := len(serviceKey)

    setNumber(memory[start:start + 2], uint16(version))
    setNumber(memory[start + 2:start + 4], uint16(serviceKeyLen))
    copy(memory[start + 4:start + 4 + ServiceKeyCap], serviceKey)
    setNumber(memory[start + 4 + ServiceKeyCap:start + 6 + ServiceKeyCap], uint16(confCnt))

    for i := uint(0);i < confCnt;i++ {
        setOneConf(memory, idx, i, confKeys[i], values[i])
    }
}

//删除某service的配置
func RemoveServiceConf(memory *[TotalConfMemSize]byte, idx uint) {
    start := idx * OneBucketCap
    //将版本号设置为0，表示此处没有配置了
    setNumber(memory[start:start + 2], 0)
}
