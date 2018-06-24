package core

import (
	"encoding/binary"
)

//获取某bucket的长度
func getBucketLen(bucket *[BucketSize]byte) uint {
	return uint(binary.LittleEndian.Uint16(bucket[:2]))
}

//为某bucket增加1长度
func addBucketLen(bucket *[BucketSize]byte) {
	nowLen := binary.LittleEndian.Uint16(bucket[:2])
	binary.LittleEndian.PutUint16(bucket[:2], nowLen + 1)
}

//为某bucket减少1长度
func decBucketLen(bucket *[BucketSize]byte) {
	nowLen := binary.LittleEndian.Uint16(bucket[:2])
	if nowLen > 0 {
		binary.LittleEndian.PutUint16(bucket[:2], nowLen - 1)
	}
}

//从某bucket中，取出索引为index的配置key
func getBucketKey(bucket *[BucketSize]byte, index uint) string {
	start := 2 + KVCap * index
	keyLen := uint(binary.LittleEndian.Uint16(bucket[start:start + 2]))
	return string(bucket[start + 2:start + 2 + keyLen])
}

//获取一个bucket中所有配置
func getBucketConfigs(bucket *[BucketSize]byte) map[string]string {
	configs := make(map[string]string)
	num := getBucketLen(bucket)
	for i := uint(0);i < num; i++ {
		key := getBucketKey(bucket, i)
		value := getBucketValue(bucket, i)
		configs[key] = value
	}
	return configs
}

//从某bucket中，取出索引为index的配置value
func getBucketValue(bucket *[BucketSize]byte, index uint) string {
	start := 2 + KVCap * index + KeyCap
	valueLen := uint(binary.LittleEndian.Uint16(bucket[start:start + 2]))
	return string(bucket[start + 2:start + 2 + valueLen])
}

//在某bucket中，设置索引为index的配置key
func setBucketKey(bucket *[BucketSize]byte, index uint, key string) {
	start := 2 + KVCap * index
	keyLen := uint16(len(key))
	//write new length
	binary.LittleEndian.PutUint16(bucket[start:start + 2], keyLen)
	//write new key
	copy(bucket[start + 2:start + 2 + KeyCap], key)
}

//在某bucket中，设置索引为index的配置value
func setBucketValue(bucket *[BucketSize]byte, index uint, value string) {
	start := 2 + KVCap * index + KeyCap
	valueLen := uint16(len(value))
	//write new length
	binary.LittleEndian.PutUint16(bucket[start:start + 2], valueLen)
	//write new key
	copy(bucket[start + 2:start + 2 + ValueCap], value)
}

//从某bucket利用二分搜索中查找是否存在某key
func searchKey(bucket *[BucketSize]byte, key string) int {
	num := getBucketLen(bucket)
	if num == 0 {
		return -1
	}
	var low uint = 0
	var high = num - 1

	for low <= high {
		mid := (low + high) / 2
		mKey := getBucketKey(bucket, mid)
		if mKey > key {
			high = mid - 1
		} else if mKey < key {
			low = mid + 1
		} else {
			//exist
			return int(mid)
		}
	}
	return -1
}

//在某bucket中利用二分搜索找到第一个字典序大于key的配置key，返回其位置
//-1表示没有这样的配置key
func firstGreat(bucket *[BucketSize]byte, key string) (int, bool) {
	num := getBucketLen(bucket)
	if num == 0 {
		return -1, false
	}
	var low uint = 0
	var high = num - 1
	for low <= high {
		mid := (low + high) / 2
		mKey := getBucketKey(bucket, mid)
		if mKey > key {
			if mid == 0 {
				return 0, false
			}
			if getBucketKey(bucket, mid - 1) < key {
				return int(mid), false
			} else {
				high = mid - 1
			}
		} else if mKey < key {
			low = mid + 1
		} else {
			//exist
			return int(mid), true
		}
	}
	return -1, false
}

//设置某配置
func setConfig(buckets *[BucketCap][BucketSize]byte, buckIdx uint, key string, value string) int {
	first, ok := firstGreat(&buckets[buckIdx], key)
	if ok {
		//key已经存在了，且索引是first，重新赋值即可
		setBucketValue(&buckets[buckIdx], uint(first), value)
	} else {
		//则需要添加key
		num := getBucketLen(&buckets[buckIdx])
		if num == BucketKVCap {
			//此配置对应的bucket已经满了
			return -1
		}
		if first == -1 {
			//不存在比key大的配置，于是准备将key添加到尾部，索引是num
			setBucketKey(&buckets[buckIdx], num, key)
			setBucketValue(&buckets[buckIdx], num, value)
		} else {
			//将新配置添加到位置first
			//需要先将此位置以及其后的配置平移，需要平移的配置有num-first个
			start := 2 + KVCap * uint(first)
			copy(buckets[buckIdx][start + KVCap:start + KVCap * (num + 1 - uint(first))],
				buckets[buckIdx][start:start + KVCap * (num - uint(first))])
			//append to first
			setBucketKey(&buckets[buckIdx], uint(first), key)
			setBucketValue(&buckets[buckIdx], uint(first), value)
		}
		//更新bucket长度
		addBucketLen(&buckets[buckIdx])
	}
	return 0
}

//获取某配置
func getConfig(buckets *[BucketCap][BucketSize]byte, buckIdx uint, key string) (string, int) {
	index := searchKey(&buckets[buckIdx], key)
	if index == -1 {
		return "", -1
	}
	return getBucketValue(&buckets[buckIdx], uint(index)), 0
}

//删除某配置
func removeConfig(buckets *[BucketCap][BucketSize]byte, buckIdx uint, key string) int {
	index := searchKey(&buckets[buckIdx], key)
	if index == -1 {
		return -1
	}
	//存在此配置
	num := getBucketLen(&buckets[buckIdx])
	if uint(index) < num - 1 {
		//此配置不是最后一项
		//则将其后的配置都前进一位
		start := 2 + KVCap * uint(index + 1)
		copy(buckets[buckIdx][start:start + KVCap * (num - uint(index))],
			buckets[buckIdx][start + KVCap:start + KVCap * (num + 1 - uint(index))])
	}
	//更新bucket长度
	decBucketLen(&buckets[buckIdx])
	return 0
}