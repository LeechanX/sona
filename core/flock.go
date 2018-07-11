package core

import (
    "syscall"
    "os"
    "io"
)

type RdFlock struct {
    file *os.File
    flock syscall.Flock_t
}

type WrFlock struct {
    file *os.File
    flock syscall.Flock_t
}

func getRdFlock(flockPath string) (*RdFlock, error) {
    fp, err := os.OpenFile(flockPath, os.O_RDONLY, 0400)
    if err != nil {
        return nil, err
    }
    flock := RdFlock {
        file: fp,
        flock: syscall.Flock_t {
            //lock for read lock
            Type:   syscall.F_RDLCK,
            Whence: io.SeekStart,
            Start:  0,
            Len:    0,
        },
    }
    return &flock, nil
}

func getWrFlock(flockPath string) (*WrFlock, error) {
    fp, err := os.OpenFile(flockPath, os.O_CREATE | os.O_RDWR, 0666)
    if err != nil {
        return nil, err
    }
    flock := WrFlock {
        file: fp,
        flock: syscall.Flock_t {
            //lock for write lock
            Type:   syscall.F_WRLCK,
            Whence: io.SeekStart,
            Start:  0,
            Len:    0,
        },
    }
    return &flock, nil
}

//上共享锁
func (rl *RdFlock) RDLock() error {
    rl.flock.Type = syscall.F_RDLCK
    return syscall.FcntlFlock(rl.file.Fd(), syscall.F_SETLKW, &(rl.flock))
}

func (rl *RdFlock) Release() error {
    rl.flock.Type = syscall.F_UNLCK
    return syscall.FcntlFlock(rl.file.Fd(), syscall.F_SETLK, &(rl.flock))
}

func (rl *RdFlock) Close() {
    rl.file.Close()
}

//可以上共享锁
func (wl *WrFlock) RDLock() error {
    wl.flock.Type = syscall.F_WRLCK
    return syscall.FcntlFlock(wl.file.Fd(), syscall.F_SETLKW, &(wl.flock))
}

//上互斥锁
func (wl *WrFlock) WRLock() error {
    wl.flock.Type = syscall.F_WRLCK
    return syscall.FcntlFlock(wl.file.Fd(), syscall.F_SETLKW, &(wl.flock))
}

//上互斥锁（非阻塞）
func (wl *WrFlock) WRLockNoWait() error {
    wl.flock.Type = syscall.F_WRLCK
    return syscall.FcntlFlock(wl.file.Fd(), syscall.F_SETLK, &(wl.flock))
}

func (wl *WrFlock) Release() error {
    wl.flock.Type = syscall.F_UNLCK
    return syscall.FcntlFlock(wl.file.Fd(), syscall.F_SETLK, &(wl.flock))
}

func (wl *WrFlock) Close() {
    wl.file.Close()
}
