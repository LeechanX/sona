package logic

import "sync"

type EditControl struct {
    keys map[string]bool//保存正在编辑的service key，用于保护数据
    mutex sync.Mutex
}

var EditingControl EditControl//记录正在编辑

//标记某service key处于“编辑”状态
//若已在编辑中，返回false
//否则返回true
func (editControl *EditControl) MarkEditing(serviceKey string) bool {
    editControl.mutex.Lock()
    defer editControl.mutex.Unlock()

    if _, ok := editControl.keys[serviceKey];ok {
        //已存在，说明已在编辑
        return false
    }
    editControl.keys[serviceKey] = true
    return true
}

//完成编辑
func (editControl *EditControl) DoneEditing(serviceKey string) {
    editControl.mutex.Lock()
    defer editControl.mutex.Unlock()

    delete(editControl.keys, serviceKey)
}
