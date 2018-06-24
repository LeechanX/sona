package core

import (
	"os"
	"syscall"
)

type SharedMem struct {
	fp *os.File
	bs []byte
}

//与mmap共享内存建立联系（没有就创建）
func attachSharedMem(filePath string, length int) (*SharedMem, error) {
	mmapFile, err := os.OpenFile(filePath, os.O_CREATE | os.O_RDWR | os.O_EXCL, 0666)
	if err != nil {
		//already created
		if os.IsExist(err) {
			mmapFile, err = os.OpenFile(filePath, os.O_RDWR, 0666)
		} else {
			return nil, err
		}
	} else {
		//create
		_, err = mmapFile.Seek(int64(length - 1), 0)
		if err != nil {
			return nil, err
		}
		_, err = mmapFile.Write([]byte(" "))
	}

	if err != nil {
		return nil, err
	}
	mmap, err := syscall.Mmap(int(mmapFile.Fd()), 0, length, syscall.PROT_READ | syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	m := SharedMem {
		fp: mmapFile,
		bs: mmap,
	}
	return &m, nil
}

//与mmap取消关联
func (shm *SharedMem) Close() {
	syscall.Munmap(shm.bs)
	shm.fp.Close()
}
