package core

import (
    "os"
    "fmt"
    "unsafe"
    "errors"
    "sona/common"
)

const (
    RootPath = "/tmp/sona"
)

//配置管理
type ConfigController struct {
    indexShm *SharedMem
    confShm *SharedMem
    indexHub *[TotalIndexMemSize]byte
    indexLock *WrFlock
    confHub *[TotalConfMemSize]byte
    confLocks [ServiceBucketLimit]*WrFlock//管理services
    gLock *WrFlock//用于确保只有一个config-controller在本机运行
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
    //获取index文件锁
    flockPath = fmt.Sprintf("%s/idx.lock", RootPath)
    iFlk, err := getWrFlock(flockPath)
    if err != nil {
        //关闭所有已打开文件锁
        for i := uint(0);i < ServiceBucketLimit; i++ {
            configController.confLocks[i].Close()
        }
        return nil, err
    }
    configController.indexLock = iFlk

    //获取index共享内存
    mapPath := fmt.Sprintf("%s/idx.mmap", RootPath)
    m, err := attachSharedMem(mapPath, int(TotalConfMemSize))
    if err != nil {
        //关闭所有文件锁
        for i := uint(0);i < ServiceBucketLimit; i++ {
            configController.confLocks[i].Close()
        }
        configController.indexLock.Close()
        return nil, err
    }
    configController.indexShm = m
    //读取共享内存，转化为数组
    configController.indexHub = (*[TotalIndexMemSize]byte)(
        unsafe.Pointer(&configController.indexShm.bs[0]))

    //获取conf共享内存
    mapPath = fmt.Sprintf("%s/cfg.mmap", RootPath)
    m, err = attachSharedMem(mapPath, int(TotalConfMemSize))
    if err != nil {
        //关闭所有文件锁
        for i := uint(0);i < ServiceBucketLimit; i++ {
            configController.confLocks[i].Close()
        }
        configController.indexLock.Close()
        return nil, err
    }
    configController.confShm = m
    //读取共享内存，转化为数组
    configController.confHub = (*[TotalConfMemSize]byte)(
        unsafe.Pointer(&configController.confShm.bs[0]))
    //已创建成功
    return &configController, nil
}

//清理
func (cc *ConfigController) Close() {
    //关闭共享内存
    cc.indexShm.Close()
    cc.confShm.Close()
    //关闭所有文件锁
    for i := uint(0);i < ServiceBucketLimit; i++ {
        cc.confLocks[i].Close()
    }
    cc.indexLock.Close()
    cc.gLock.Release()
    cc.gLock.Close()
}

//获取当前所有serviceKey:版本
func (cc *ConfigController) GetAllServiceKeys() map[string]uint {
    cc.indexLock.RDLock()
    defer cc.indexLock.Release()
    return GetAllService(cc.indexHub)
}

//某serviceKey是否存在
func (cc *ConfigController) IsServiceExist(serviceKey string) bool {
    cc.indexLock.RDLock()
    defer cc.indexLock.Release()
    ret, _ := GetServiceIndex(cc.indexHub, serviceKey)
    return ret != -1
}

/*
 * 修改Tip：再操作具体conf，再更新index里的版本
*/

//新增一个service配置
func (cc *ConfigController) addNewService(serviceKey string, remoteVersion uint, configKeys []string, values []string) error {
    index := GetFirstIndexFree(cc.indexHub)
    if index == ServiceBucketLimit {
        //已经没空间了
        return errors.New("no more space to store new service configures")
    }
    //设置conf
    //先排序
    sortedKeys, sortedValues := common.SortKV(configKeys, values)
    cc.confLocks[index].WRLock()
    AddServiceConf(cc.confHub, uint(index), sortedKeys, sortedValues)
    cc.confLocks[index].Release()
    //新增到index hub中
    cc.indexLock.WRLock()
    InsertService(cc.indexHub, serviceKey, remoteVersion, index)
    cc.indexLock.Release()
    return nil
}
/*
//写配置：为某serviceKey新增一个、修改一个
func (cc *ConfigController) Set(serviceKey string, configKey string, value string, remoteVersion uint) error {
    //先获取索引位置
    cc.indexLock.RDLock()
    index, localVersion := GetServiceIndex(cc.indexHub, serviceKey)
    cc.indexLock.Release()
    if index == -1 {
        //则需要新增service
        return cc.addNewService(serviceKey, remoteVersion, []string{configKey}, []string{value})
    }
    if remoteVersion <= localVersion {
        return errors.New(fmt.Sprintf("remote service %s configure is too old", serviceKey))
    }
    //更新
    cc.confLocks[index].WRLock()
    defer cc.confLocks[index].Release()
    if !AddOrUpdateConf(cc.confHub, uint(index), configKey, value) {
        return errors.New("no more space to store new service configures")
    }
    return nil
}
*/
/*
//删除一个配置
func (cc *ConfigController) RemoveOne(serviceKey string, configKey string, remoteVersion uint) {
    //先获取索引位置
    cc.indexLock.RDLock()
    index, localVersion := GetServiceIndex(cc.indexHub, serviceKey)
    cc.indexLock.Release()

    if index == -1 || remoteVersion <= localVersion{
        return
    }

    //删除这一项
    cc.confLocks[index].WRLock()
    RemoveOneConf(cc.confHub, uint(index), configKey)
    confCnt := GetConfCount(cc.confHub, uint(index))
    cc.confLocks[index].Release()

    cc.indexLock.WRLock()
    defer cc.indexLock.Release()
    if confCnt == 0 {
        //此service已空，删除
        RemoveService(cc.indexHub, serviceKey)
    } else {
        //更新最新版本
        UpdServiceVersion(cc.indexHub, serviceKey, remoteVersion)
    }
}
*/

//删除某service的配置
func (cc *ConfigController) RemoveService(serviceKey string) {
    //先获取索引位置
    cc.indexLock.WRLock()
    index, _ := GetServiceIndex(cc.indexHub, serviceKey)
    if index != -1 {
        RemoveService(cc.indexHub, serviceKey)
    }
    cc.indexLock.Release()
    if index != -1 {
        cc.confLocks[index].WRLock()
        RemoveServiceConf(cc.confHub, uint(index))
        cc.confLocks[index].Release()
    }
}

//更新一个service的配置，来自于pull的回复
func (cc *ConfigController) UpdateService(serviceKey string, remoteVersion uint, configKeys []string, values []string) error {
    cc.indexLock.RDLock()
    index, localVersion := GetServiceIndex(cc.indexHub, serviceKey)
    cc.indexLock.Release()
    if index == -1 {
        //说明本地不存在此serviceKey，那么添加
        return cc.addNewService(serviceKey, remoteVersion, configKeys, values)
    }
    if remoteVersion <= localVersion {
        return errors.New(fmt.Sprintf("remote service %s configure is too old", serviceKey))
    }
    //更新版本
    cc.indexLock.WRLock()
    UpdServiceVersion(cc.indexHub, serviceKey, remoteVersion)
    cc.indexLock.Release()
    //先排序
    sortedKeys, sortedValues := common.SortKV(configKeys, values)
    cc.confLocks[index].WRLock()
    if len(configKeys) == 0 {
        //说明远端删除了整个service
        RemoveServiceConf(cc.confHub, uint(index))
    } else {
        //否则，重置新的serviceConf
        AddServiceConf(cc.confHub, uint(index), sortedKeys, sortedValues)
    }
    cc.confLocks[index].Release()
    return nil
}

//读配置
func (cc *ConfigController) Get(serviceKey string, confKey string) string {
    cc.indexLock.RDLock()
    index, _ := GetServiceIndex(cc.indexHub, serviceKey)
    cc.indexLock.Release()
    if index == -1 {
        //不存在
        return ""
    }
    //上读锁
    cc.confLocks[index].RDLock()
    defer cc.confLocks[index].Release()
    return GetConf(cc.confHub, uint(index), confKey)
}

//配置获取
type ConfigGetter struct {
    indexShm *SharedMem
    confShm *SharedMem
    indexHub *[TotalIndexMemSize]byte
    indexLock *RdFlock
    confHub *[TotalConfMemSize]byte
    confLocks [ServiceBucketLimit]*RdFlock//管理services
}

//创建一个配置获取
func GetConfigGetter() (*ConfigGetter, error) {
    configGetter := ConfigGetter {}
    //获取文件锁
    for i := uint(0);i < ServiceBucketLimit; i++ {
        flockPath := fmt.Sprintf("%s/cfg_%d.lock", RootPath, i)
        fl, err := getRdFlock(flockPath)
        if err != nil {
            //关闭所有已打开文件锁
            for j := uint(0);j < i; j++ {
                configGetter.confLocks[j].Close()
            }
            return nil, err
        }
        configGetter.confLocks[i] = fl
    }
    //获取index文件锁
    flockPath := fmt.Sprintf("%s/idx.lock", RootPath)
    iFlk, err := getRdFlock(flockPath)
    if err != nil {
        //关闭所有已打开文件锁
        for i := uint(0);i < ServiceBucketLimit; i++ {
            configGetter.confLocks[i].Close()
        }
        return nil, err
    }
    configGetter.indexLock = iFlk

    //获取index共享内存
    mapPath := fmt.Sprintf("%s/idx.mmap", RootPath)
    m, err := attachSharedMem(mapPath, int(TotalConfMemSize))
    if err != nil {
        //关闭所有文件锁
        for i := uint(0);i < ServiceBucketLimit; i++ {
            configGetter.confLocks[i].Close()
        }
        configGetter.indexLock.Close()
        return nil, err
    }
    configGetter.indexShm = m
    //读取共享内存，转化为数组
    configGetter.indexHub = (*[TotalIndexMemSize]byte)(
        unsafe.Pointer(&configGetter.indexShm.bs[0]))

    //获取conf共享内存
    mapPath = fmt.Sprintf("%s/cfg.mmap", RootPath)
    m, err = attachSharedMem(mapPath, int(TotalConfMemSize))
    if err != nil {
        //关闭所有文件锁
        for i := uint(0);i < ServiceBucketLimit; i++ {
            configGetter.confLocks[i].Close()
        }
        configGetter.indexLock.Close()
        return nil, err
    }
    configGetter.confShm = m
    //读取共享内存，转化为数组
    configGetter.confHub = (*[TotalConfMemSize]byte)(
        unsafe.Pointer(&configGetter.confShm.bs[0]))
    //已创建成功
    return &configGetter, nil
}

//清理
func (cg *ConfigGetter) Close() {
    //关闭共享内存
    cg.indexShm.Close()
    cg.confShm.Close()
    //关闭所有文件锁
    for i := uint(0);i < ServiceBucketLimit; i++ {
        cg.confLocks[i].Close()
    }
    cg.indexLock.Close()
}

//读配置
func (cg *ConfigGetter) Get(serviceKey string, confKey string) string {
    cg.indexLock.RDLock()
    index, _ := GetServiceIndex(cg.indexHub, serviceKey)
    cg.indexLock.Release()
    if index == -1 {
        //不存在
        return ""
    }
    //上读锁
    cg.confLocks[index].RDLock()
    defer cg.confLocks[index].Release()
    return GetConf(cg.confHub, uint(index), confKey)
}
