package core

import (
	"os"
	"fmt"
	"unsafe"
	"errors"
	"github.com/gwenn/murmurhash3"
	"sync"
)

const (
	FieldNumber = 5
	//产品线名.业务组名.服务名.所需section.配置key, 其中“产品线名.业务组名.服务名”组成serviceKey用于标识一个服务
	//每个字段不得超过30字节
	KeyCap uint = 155
	ValueCap uint = 35
	//bucket个数
	//多出一个bucket用于存放那些其bucket已装不下的配置们
	BucketCap uint = 100 + 1
	//单个bucket中存放多少个KV
	BucketKVCap uint = 200

	//2 byte length(uint16) for key, 2 byte length(uint16) value
	KVCap = 2 + KeyCap + 2 + ValueCap
	//2 byte for number of kv
	BucketSize = 2 + KVCap * BucketKVCap

	RootPath = "/tmp/easyconfig"
)

//配置管理
type ConfigController struct {
	shm *SharedMem
	buckets *[BucketCap][BucketSize]byte
	locks [BucketCap]*WrFlock//管理buckets
	totalConfigs map[string]map[string]string//将配置也用map形式存在agent自己的内存上，方便更新时对比
	//格式：
	//serviceKey1: configKey1:configValue1, configKey2:configValue2...
	//serviceKey2: configKey1:configValue1, configKey2:configValue2...
	mutex sync.Mutex//保护map形式的配置

	gLock *WrFlock//用于确保只有一个agent在本机运行
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
	configController := ConfigController {}
	//先确定是否本机仅有一个controller
	flockPath := fmt.Sprintf("%s/global_flock.flk", RootPath)
	gfl, err := getWrFlock(flockPath)
	if err != nil {
		return nil, err
	}
	if err = gfl.WRLockNoWait();err != nil {
		//说明已有agent进程存在了
		return nil, err
	}
	//agent运行期间，将一直对global_flock.flk上互斥锁
	configController.gLock = gfl

	//获取文件锁
	for i := uint(0);i < BucketCap; i++ {
		flockPath := fmt.Sprintf("%s/flock_%d.flk", RootPath, i)
		fl, err := getWrFlock(flockPath)
		if err != nil {
			//关闭所有已打开文件锁
			for j := uint(0);j < i; j++ {
				configController.locks[j].Close()
			}
			return nil, err
		}
		configController.locks[i] = fl
	}
	//获取共享内存
	mmapPath := fmt.Sprintf("%s/mmap.mp", RootPath)
	m, err := attachSharedMem(mmapPath, int(BucketSize * BucketCap))
	if err != nil {
		//关闭所有文件锁
		for i := uint(0);i < BucketCap; i++ {
			configController.locks[i].Close()
		}
		return nil, err
	}
	configController.shm = m
	//读取共享内存，转化为数组
	configController.buckets = (*[BucketCap][BucketSize]byte)(unsafe.Pointer(&m.bs[0]))
	//已创建成功
	//加载所有配置项到agent内存里，方便以后更新使用
	configController.loadAll()
	return &configController, nil
}

//加载所有配置项到agent内存里，方便以后更新使用
//仅在启动时执行
func (cc *ConfigController) loadAll() {
	cc.totalConfigs = make(map[string]map[string]string)
	for idx := uint(0);idx < BucketCap; idx++ {
		cc.locks[idx].RDLock()
		bucketTotalConfigs := getBucketTotalConfigs(&cc.buckets[idx])
		cc.locks[idx].Release()
		for key, value := range bucketTotalConfigs {
			serviceKey := GetServiceKey(key)
			configKey := GetConfigKey(key)
			if _, ok := cc.totalConfigs[serviceKey];!ok {
				cc.totalConfigs[serviceKey] = make(map[string]string)
			}
			cc.totalConfigs[serviceKey][configKey] = value
		}
	}
}

//清理
func (cc *ConfigController) Close() {
	//关闭共享内存
	cc.shm.Close()
	//关闭所有文件锁
	for i := uint(0);i < BucketCap; i++ {
		cc.locks[i].Close()
	}
	cc.gLock.Release()
	cc.gLock.Close()
}

//获取当前所有serviceKey
func (cc *ConfigController) GetAllServiceKeys() map[string]bool {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	result := make(map[string]bool)
	for key := range cc.totalConfigs {
		result[key] = true
	}
	return result
}

//某serviceKey是否存在
func (cc *ConfigController) ExistService(serviceKey string) bool {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	_, ok := cc.totalConfigs[serviceKey]
	return ok
}

//写配置：新增、修改
func (cc *ConfigController) Set(serviceKey string, configKey string, value string) error {
	key := SpliceKey(serviceKey, configKey)
	//check valid item
	if !IsValidityKey(key) || value == "" {
		return errors.New("empty key/value")
	}
	keyLen := len(key)
	valueLen := len(value)
	//check if key or value's length is too large
	if uint(keyLen) > KeyCap {
		return errors.New(fmt.Sprintf("key's length is too large, limit is %d", KeyCap))
	}
	if uint(valueLen) > ValueCap {
		return errors.New(fmt.Sprintf("value's length is too large, limit is %d", ValueCap))
	}
	buckIndex := uint(murmurhash3.Murmur3A([]byte(key), 0)) % (BucketCap - 1)
	//上互斥锁
	cc.locks[buckIndex].WRLock()
	ret := setConfig(cc.buckets, buckIndex, key, value)
	cc.locks[buckIndex].Release()
	if ret == -1 {
		//对应bucket已经满了
		//则尝试放到最后一个bucket
		cc.locks[BucketCap - 1].WRLock()
		defer cc.locks[BucketCap - 1].Release()
		ret = setConfig(cc.buckets, BucketCap - 1, key, value)
		if ret == -1 {
			return errors.New("configure hub is full")
		}
	}
	//同时更新本地map配置
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	if _, ok := cc.totalConfigs[serviceKey];!ok {
		cc.totalConfigs[serviceKey] = make(map[string]string)
	}
	cc.totalConfigs[serviceKey][configKey] = value
	return nil
}

//删除一个配置
func (cc *ConfigController) RemoveOne(serviceKey string, configKey string) {
	key := SpliceKey(serviceKey, configKey)
	//check valid item
	if !IsValidityKey(key) {
		return
	}
	buckIndex := uint(murmurhash3.Murmur3A([]byte(key), 0)) % (BucketCap - 1)
	//上互斥锁
	cc.locks[buckIndex].WRLock()
	ret := removeConfig(cc.buckets, buckIndex, key)
	cc.locks[buckIndex].Release()
	if ret == -1 {
		//对应bucket不存在此项
		//则尝试在最后一个bucket中删除此项
		cc.locks[BucketCap - 1].WRLock()
		_ = removeConfig(cc.buckets, BucketCap - 1, key)
		cc.locks[BucketCap - 1].Release()

		//顺便在map配置里删除之
		cc.mutex.Lock()
		defer cc.mutex.Unlock()
		if _, ok := cc.totalConfigs[serviceKey];ok {
			delete(cc.totalConfigs[serviceKey], configKey)
			if len(cc.totalConfigs[serviceKey]) == 0 {
				delete(cc.totalConfigs, serviceKey)
			}
		}
	}
}

//删除一个service的配置
func (cc *ConfigController) Remove(serviceKey string) {
	cc.mutex.Lock()
	serviceConfigs, ok := cc.totalConfigs[serviceKey]
	if !ok {
		//压根不存在
		cc.mutex.Unlock()
		return
	}
	//在map配置中删除
	delete(cc.totalConfigs, serviceKey)
	cc.mutex.Unlock()

	//在共享内存中删除
	for configKey := range serviceConfigs {
		key := SpliceKey(serviceKey, configKey)
		buckIndex := uint(murmurhash3.Murmur3A([]byte(key), 0)) % (BucketCap - 1)
		//上互斥锁
		cc.locks[buckIndex].WRLock()
		ret := removeConfig(cc.buckets, buckIndex, key)
		cc.locks[buckIndex].Release()
		if ret == -1 {
			//对应bucket不存在此项
			//则尝试在最后一个bucket中删除此项
			cc.locks[BucketCap - 1].WRLock()
			_ = removeConfig(cc.buckets, BucketCap - 1, key)
			cc.locks[BucketCap - 1].Release()
		}
	}
}

//更新一个service的配置
func (cc *ConfigController) UpdateService(serviceKey string, configKeys []string, values []string) {
	if len(configKeys) != len(values) {
		return
	}
	//remote config
	remoteServiceConfigs := make(map[string]string)
	for i := range configKeys {
		remoteServiceConfigs[configKeys[i]] = values[i]
	}
	cc.mutex.Lock()
	serviceConfigs, ok := cc.totalConfigs[serviceKey]
	if !ok {
		//压根不存在
		cc.mutex.Unlock()
		return
	}
	cc.mutex.Unlock()
	//获取哪些需要新增、更新
	for key, value := range remoteServiceConfigs {
		if _, ok := serviceConfigs[key];!ok {
			//是新配置，添加
			cc.Set(serviceKey, key, value)
		} else if serviceConfigs[key] != value {
			//是更新的值，更新
			cc.Set(serviceKey, key, value)
		}
	}
	//哪些需要删除
	for key := range serviceConfigs {
		if _, ok := remoteServiceConfigs[key];!ok {
			//被删除了
			cc.RemoveOne(serviceKey, key)
		}
	}
}

//读配置
func (cc *ConfigController) Get(key string) (string, error) {
	//check empty item
	if key == "" {
		return "", errors.New("empty key")
	}
	buckIndex := uint(murmurhash3.Murmur3A([]byte(key), 0)) % (BucketCap - 1)
	//上读锁
	cc.locks[buckIndex].RDLock()
	value, ret := getConfig(cc.buckets, buckIndex, key)
	cc.locks[buckIndex].Release()
	if ret == -1 {
		//不存在，可能在最后一个bucket
		cc.locks[BucketCap - 1].RDLock()
		defer cc.locks[BucketCap - 1].Release()
		value, ret = getConfig(cc.buckets, BucketCap - 1, key)
		if ret == -1 {
			return "", errors.New("configure is not exist")
		} else {
			return value, nil
		}
	}
	return value, nil
}

//配置总数
func (cc *ConfigController) GetSize() uint {
	var number uint
	for i := uint(0);i < BucketCap; i++ {
		//上读锁
		cc.locks[i].RDLock()
		number += getBucketLen(&cc.buckets[i])
		cc.locks[i].Release()
	}
	return number
}

//配置获取
type ConfigGetter struct {
	shm *SharedMem
	buckets *[BucketCap][BucketSize]byte
	locks [BucketCap]*RdFlock
}

//创建一个配置获取
func GetConfigGetter() (*ConfigGetter, error) {
	configGetter := ConfigGetter {}
	//获取文件锁
	for i := uint(0);i < BucketCap; i++ {
		flockPath := fmt.Sprintf("%s/flock_%d.flk", RootPath, i)
		fl, err := getRdFlock(flockPath)
		if err != nil {
			//关闭所有已打开文件锁
			for j := uint(0);j < i; j++ {
				configGetter.locks[j].Close()
			}
			return nil, err
		}
		configGetter.locks[i] = fl
	}
	mmapPath := fmt.Sprintf("%s/mmap.mp", RootPath)
	m, err := attachSharedMem(mmapPath, int(BucketSize * BucketCap))
	if err != nil {
		//关闭文件锁
		for i := uint(0);i < BucketCap; i++ {
			configGetter.locks[i].Close()
		}
		return nil, err
	}
	configGetter.shm = m
	//读取共享内存，转化为数组
	configGetter.buckets = (*[BucketCap][BucketSize]byte)(unsafe.Pointer(&m.bs[0]))
	return &configGetter, nil
}

//清理
func (cg *ConfigGetter) Close() {
	//关闭共享内存
	cg.shm.Close()
	//关闭所有文件锁
	for i := uint(0);i < BucketCap; i++ {
		cg.locks[i].Close()
	}
}

//读配置
func (cg *ConfigGetter) Get(key string) (string, error) {
	//check empty item
	if key == "" {
		return "", errors.New("empty key")
	}
	buckIndex := uint(murmurhash3.Murmur3A([]byte(key), 0)) % (BucketCap - 1)
	//上读锁
	cg.locks[buckIndex].RDLock()
	value, ret := getConfig(cg.buckets, buckIndex, key)
	cg.locks[buckIndex].Release()
	if ret == -1 {
		//不存在，可能在最后一个bucket
		cg.locks[BucketCap - 1].RDLock()
		defer cg.locks[BucketCap - 1].Release()
		value, ret = getConfig(cg.buckets, BucketCap - 1, key)
		if ret == -1 {
			return "", errors.New("configure is not exist")
		} else {
			return value, nil
		}
	}
	return value, nil
}

//配置总数
func (cg *ConfigGetter) GetSize() uint {
	//上读锁
	var number uint
	for i := uint(0);i < BucketCap; i++ {
		//上读锁
		cg.locks[i].RDLock()
		number += getBucketLen(&cg.buckets[i])
		cg.locks[i].Release()
	}
	return number
}
