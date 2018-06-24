package core

import (
	"os"
	"fmt"
	"unsafe"
	"errors"
	"github.com/gwenn/murmurhash3"
)

const (
	KeyCap uint = 160
	ValueCap uint = 40
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
	locks [BucketCap]*WrFlock
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
	return &configController, nil
}

//清理
func (cc *ConfigController) Close() {
	//关闭共享内存
	cc.shm.Close()
	//关闭所有文件锁
	for i := uint(0);i < BucketCap; i++ {
		cc.locks[i].Close()
	}
}

//写配置
func (cc *ConfigController) Set(key string, value string) error {
	//check empty item
	if key == "" || value == "" {
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
		} else {
			return nil
		}
	}
	return nil
}

//删除配置
func (cc *ConfigController) Remove(key string) {
	if key == "" {
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
		defer cc.locks[BucketCap - 1].Release()
		_ = removeConfig(cc.buckets, BucketCap - 1, key)
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

//获取一个bucket所有配置
func (cc *ConfigController) GetAll(idx uint) map[string]string {
	cc.locks[idx].RDLock()
	defer cc.locks[idx].Release()
	return getBucketConfigs(&cc.buckets[idx])
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
