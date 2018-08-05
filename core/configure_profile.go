package core

import (
    "os"
    "fmt"
    "sync"
    "unsafe"
    "errors"
    "sona/common"
)

const (
    RootPath = "/tmp/sona"
)

//配置管理
type ConfigController struct {
    sharedMemory *SharedMem
    confMemory *[TotalConfMemSize]byte
    confLocks [ServiceBucketLimit]*WrFlock//管理services
    gLock *WrFlock//用于确保只有一个config-controller在本机运行

    indexMap map[string]uint//位置管理,service Key=>position
    indexLock sync.RWMutex//保护索引map
}

//创建一个配合控制
func GetConfigController() (*ConfigController, error) {
    //先尝试创建目录
    _, err := os.Stat(RootPath)
    if err != nil {
        if os.IsNotExist(err) {
            //不存在，创建目录
            err = os.Mkdir(RootPath, 0777)
            if err != nil {
                //创建出错
                return nil, err
            }
        } else {
            return nil, err
        }
    }
    configController := ConfigController{}
    //先确定是否本机仅有一个controller
    flockPath := fmt.Sprintf("%s/global.lock", RootPath)
    gFlk, err := getWrFlock(flockPath)
    if err != nil {
        return nil, err
    }
    if err = gFlk.WRLockNoWait();err != nil {
        //说明已有agent进程存在了
        return nil, errors.New("local machine is already running a sona-agent")
    }
    //config-controller运行期间，将一直对global_flock.flk上互斥锁
    configController.gLock = gFlk

    //获取conf文件锁
    for i := uint(0);i < ServiceBucketLimit; i++ {
        flockPath := fmt.Sprintf("%s/cfg_%d.lock", RootPath, i)
        cFlk, err := getWrFlock(flockPath)
        if err != nil {
            //关闭所有已打开文件锁
            for j := uint(0);j < i; j++ {
                configController.confLocks[j].Close()
            }
            return nil, err
        }
        configController.confLocks[i] = cFlk
    }

    //获取conf共享内存
    mapPath := fmt.Sprintf("%s/cfg.mmap", RootPath)
    m, err := attachSharedMem(mapPath, int(TotalConfMemSize))
    if err != nil {
        //关闭所有文件锁
        for i := uint(0);i < ServiceBucketLimit; i++ {
            configController.confLocks[i].Close()
        }
        return nil, err
    }
    configController.sharedMemory = m
    //读取共享内存，转化为数组
    configController.confMemory = (*[TotalConfMemSize]byte)(
        unsafe.Pointer(&configController.sharedMemory.bs[0]))
    //已创建成功
    //获取目前所有配置的索引
    configController.indexMap = GetAllServiceIndex(configController.confMemory)
    return &configController, nil
}

//清理
func (cc *ConfigController) Close() {
    //关闭共享内存
    cc.sharedMemory.Close()
    //关闭所有文件锁
    for i := uint(0);i < ServiceBucketLimit; i++ {
        cc.confLocks[i].Close()
    }
    cc.gLock.Release()
    cc.gLock.Close()
}

//获取当前所有serviceKey与版本
func (cc *ConfigController) GetAllServiceKeys() map[string]uint {
    keys := make(map[string]uint)
    for idx := uint(0);idx < ServiceBucketLimit; idx++ {
        cc.confLocks[idx].RDLock()
        serviceKey, version := GetServiceKey(cc.confMemory, idx)
        cc.confLocks[idx].Release()
        if version != 0 {
            keys[serviceKey] = version
        }
    }
    return keys
}

//某serviceKey是否存在
func (cc *ConfigController) IsServiceExist(serviceKey string) bool {
    cc.indexLock.RLock()
    defer cc.indexLock.RUnlock()
    _, ok := cc.indexMap[serviceKey]
    return ok
}

//获取第一个可用索引
func (cc* ConfigController) GetFirstIndexFree() uint {
    isFilled := make([]bool, ServiceBucketLimit)
    cc.indexLock.RLock()
    for _, index := range cc.indexMap {
        //此位置已被填充
        isFilled[index] = true
    }
    cc.indexLock.RUnlock()
    for index := uint(0);index < ServiceBucketLimit;index++ {
        if !isFilled[index] {
            return index
        }
    }
    return ServiceBucketLimit
}

//新增一个service配置
func (cc *ConfigController) addNewService(serviceKey string, remoteVersion uint, configKeys []string, values []string) error {
    index := cc.GetFirstIndexFree()
    if index == ServiceBucketLimit {
        //已经没空间了
        return errors.New("no more space to store new service configures")
    }
    //设置conf,先排序
    sortedKeys, sortedValues := common.SortKV(configKeys, values)
    cc.confLocks[index].WRLock()
    AddServiceConf(cc.confMemory, uint(index), serviceKey, remoteVersion, sortedKeys, sortedValues)
    cc.confLocks[index].Release()
    //更新到索引
    cc.indexLock.Lock()
    defer cc.indexLock.Unlock()
    cc.indexMap[serviceKey] = index
    return nil
}

//删除某service的配置 带版本 （broker数据到来触发）
func (cc *ConfigController) RemoveService(serviceKey string, remoteVersion uint) {
    //先获取索引位置
    cc.indexLock.Lock()
    index, ok := cc.indexMap[serviceKey]
    cc.indexLock.Unlock()
    if !ok {
        //不存在
        return
    }
    //获取版本
    _, localVersion := GetServiceKey(cc.confMemory, index)
    if localVersion >= remoteVersion {
        //远程版本太低 放弃
        return
    }

    //执行删除，先删除索引
    cc.indexLock.Lock()
    delete(cc.indexMap, serviceKey)
    cc.indexLock.Unlock()

    //在内存中删除
    cc.confLocks[index].WRLock()
    defer cc.confLocks[index].Release()
    RemoveServiceConf(cc.confMemory, index)
}

//强制删除某service的配置 无视版本（清理G触发）
func (cc *ConfigController) ForceRemoveService(serviceKey string) {
    //先获取索引位置
    cc.indexLock.Lock()
    index, ok := cc.indexMap[serviceKey]
    cc.indexLock.Unlock()
    if !ok {
        //不存在
        return
    }
    //执行删除，先删除索引
    cc.indexLock.Lock()
    delete(cc.indexMap, serviceKey)
    cc.indexLock.Unlock()

    //在内存中删除
    cc.confLocks[index].WRLock()
    defer cc.confLocks[index].Release()
    RemoveServiceConf(cc.confMemory, index)
}

//更新一个service的配置，来自于pull的回复 前提：配置数据不为空
func (cc *ConfigController) UpdateService(serviceKey string, remoteVersion uint, configKeys []string, values []string) error {
    cc.indexLock.RLock()
    index, ok := cc.indexMap[serviceKey]
    cc.indexLock.RUnlock()
    if !ok {
        //说明本地不存在此serviceKey，那么添加
        return cc.addNewService(serviceKey, remoteVersion, configKeys, values)
    }
    //获取版本
    _, localVersion := GetServiceKey(cc.confMemory, index)
    if localVersion >= remoteVersion {
        //远程版本不是新的 放弃
        return errors.New(fmt.Sprintf("remote service %s configure is old", serviceKey))
    }
    //先排序
    sortedKeys, sortedValues := common.SortKV(configKeys, values)

    //更新
    cc.confLocks[index].WRLock()
    defer cc.confLocks[index].Release()
    //重置新的serviceConf
    AddServiceConf(cc.confMemory, uint(index), serviceKey, remoteVersion, sortedKeys, sortedValues)
    return nil
}

//查询某serviceKey的索引
func (cc *ConfigController) QueryIndex(serviceKey string) (uint, error) {
    cc.indexLock.RLock()
    defer cc.indexLock.RUnlock()
    index, ok := cc.indexMap[serviceKey]
    if !ok {
        return ServiceBucketLimit, errors.New("service is not exist")
    }
    return index, nil
}

//配置获取,代表一个serviceKey
type ConfigGetter struct {
    sharedMemory *SharedMem
    confMemory *[TotalConfMemSize]byte
    index uint
    serviceKey string
    confLock *RdFlock//控制访问services
}

//创建一个配置获取
func GetConfigGetter(serviceKey string, index uint) (*ConfigGetter, error) {
    configGetter := ConfigGetter{
        serviceKey:serviceKey,
        index:index,
    }

    //获取conf文件锁
    flockPath := fmt.Sprintf("%s/cfg_%d.lock", RootPath, index)
    flock, err := getRdFlock(flockPath)
    if err != nil {
        return nil, err
    }
    configGetter.confLock = flock

    //获取conf共享内存
    mapPath := fmt.Sprintf("%s/cfg.mmap", RootPath)
    m, err := attachSharedMem(mapPath, int(TotalConfMemSize))
    if err != nil {
        return nil, err
    }
    configGetter.sharedMemory = m
    //读取共享内存，转化为数组
    configGetter.confMemory = (*[TotalConfMemSize]byte)(
        unsafe.Pointer(&configGetter.sharedMemory.bs[0]))
    //已创建成功
    return &configGetter, nil
}

//清理
func (cg *ConfigGetter) Close() {
    //关闭共享内存
    cg.sharedMemory.Close()
    //关闭文件锁
    cg.confLock.Release()
}

//读配置
func (cg *ConfigGetter) Get(confKey string) string {
    //上读锁
    cg.confLock.RDLock()
    defer cg.confLock.Release()
    return GetConf(cg.confMemory, cg.serviceKey, uint(cg.index), confKey)
}
