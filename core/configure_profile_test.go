package core

import (
    "testing"
    "fmt"
)

var configController *ConfigController
var valuePrefix string

func TestConfigController_GetAllServiceKeys(t *testing.T) {
    allServiceKeys := configController.GetAllServiceKeys()
    if len(allServiceKeys) != 99 {
        t.Fatalf("the total number of all service key: expect 99, actual %d\n", len(allServiceKeys))
    }

    for i := 0;i < 99;i++ {
        serviceKey := fmt.Sprintf("test.servicekey.key-%d", i + 1)
        version, ok := allServiceKeys[serviceKey]
        if !ok {
            t.Errorf("service key: %s expect exist, actual not\n", serviceKey)
        }
        if version != uint(i + 1) {
            t.Errorf("service version: expect %d, actual %d\n", uint(i + 1), version)
        }
    }
}

func TestConfigController_IsServiceExist(t *testing.T) {
    for i := 0;i < 99;i++ {
        serviceKey := fmt.Sprintf("test.servicekey.key-%d", i + 1)
        if !configController.IsServiceExist(serviceKey) {
            t.Errorf("service key: %s expect exist, actual not\n", serviceKey)
        }
    }

    if configController.IsServiceExist("test.servicekey.key-0") {
        t.Error("service key: test.service.key-00 expect not exist, actual exist")
    }
    if configController.IsServiceExist("test.servicekey.key-100") {
        t.Error("service key: test.service.key-100 expect not exist, actual exist")
    }
}

func TestConfigGetter_Get(t *testing.T) {
    //正常获取
    for i := 0;i < 99; i++ {
        serviceKey := fmt.Sprintf("test.servicekey.key-%d", i + 1)
        index, err := configController.QueryIndex(serviceKey)
        if err != nil {
            t.Errorf("query %s's index got error: %s\n", serviceKey, err)
        }
        if index != uint(i) {
            t.Errorf("query %s's index value expect %d, actual %d\n", serviceKey, index, i)
        }
        configGetter, err := GetConfigGetter(serviceKey, index)
        if err != nil {
            t.Errorf("create %s's getter got error: %s\n", serviceKey, err)
        }

        for j := 0;j < i + 1; j++ {
            confKey := fmt.Sprintf("service-%d.conf-%d", i, j)
            expectConfValue := fmt.Sprintf("%s-configure-value-%d", valuePrefix, j)
            value := configGetter.Get(confKey)
            if value != expectConfValue {
                t.Errorf("get service %s's %s value: expect %s actual %s\n", serviceKey, confKey, expectConfValue, value)
            }
        }
    }
    //故意获取错误的
    configGetter, err := GetConfigGetter("no_exist_service_key", 10)
    if err != nil {
        t.Errorf("create no_exist_service_key's getter got error: %s\n", err)
    }
    value := configGetter.Get("service-9.conf-0")
    if value != "" {
        t.Errorf("get service service-9.conf-0's yy.yy value expect '' actual %s", value)
    }

    configGetter, err = GetConfigGetter("test.servicekey.key-2", 1)
    if err != nil {
        t.Errorf("create test.servicekey.key-2's getter got error: %s\n", err)
    }
    value = configGetter.Get("service-??.conf-??")
    if value != "" {
        t.Errorf("get service service-??.conf-??'s value expect '' actual %s", value)
    }

    freeIndex := configController.GetFirstIndexFree()
    if freeIndex != 99 {
        t.Errorf("first free index expect 99, actual %d\n", freeIndex)
    }
}

func TestConfigController_UpdateService(t *testing.T) {
    //1：增加新
    confKeys := []string{"app.i","pap.j","ppa.k"}
    confValues := []string{"i_v","j_v","k_v"}
    err := configController.UpdateService("test.serviceKey.newadd", 1, confKeys, confValues)
    if err != nil {
        t.Error(err)
    }
    if !configController.IsServiceExist("test.serviceKey.newadd") {
        t.Error("test.serviceKey.newadd should exist")
    }
    num := len(configController.GetAllServiceKeys())
    if num != 100 {
        t.Errorf("after add, expect 100 actual %d\n", num)
    }

    freeIndex := configController.GetFirstIndexFree()
    if freeIndex != ServiceBucketLimit {
        t.Errorf("first free index expect 100, actual %d\n", freeIndex)
    }

    index, err := configController.QueryIndex("test.serviceKey.newadd")
    if index != 99 {
        t.Errorf("first free index expect 99, actual %d\n", index)
    }

    configGetter, err := GetConfigGetter("test.serviceKey.newadd", index)
    if err != nil {
        t.Errorf("create test.serviceKey.newadd's getter got error: %s\n", err)
    }

    value := configGetter.Get("app.i")
    if value != "i_v" {
        t.Errorf("expect i_v actual %s\n", value)
    }
    value = configGetter.Get("pap.j")
    if value != "j_v" {
        t.Errorf("expect j_v actual %s\n", value)
    }
    value = configGetter.Get("ppa.k")
    if value != "k_v" {
        t.Errorf("expect k_v actual %s\n", value)
    }

    //2：增加新：满了：应该失败
    err = configController.UpdateService("test.serviceKey.newadd2", 1, confKeys, confValues)
    if err == nil {
        t.Error("add should meet full error, but not")
    }

    num = len(configController.GetAllServiceKeys())
    if num != 100 {
        t.Errorf("after add, expect 100 actual %d\n", num)
    }
    if configController.IsServiceExist("test.serviceKey.newadd2") {
        t.Error("test.serviceKey.newadd2 shouldn't exist")
    }

    //3：修改: 版本对
    //test.servicekey.key-2, 2, service-1.conf-2 configure-value-2
    err = configController.UpdateService("test.servicekey.key-3", 5, []string{"aaa.bbb","ccc.ddd"}, []string{"eee","fff"})
    configGetter, err = GetConfigGetter("test.servicekey.key-3", 2)
    if err != nil {
        t.Errorf("create test.servicekey.key-3's getter got error: %s\n", err)
    }
    value = configGetter.Get("aaa.bbb")
    if value != "eee" {
        t.Errorf("expect eee actual %s\n", value)
    }

    value = configGetter.Get("ccc.ddd")
    if value != "fff" {
        t.Errorf("expect fff actual %s\n", value)
    }

    //4: 修改：版本不对
    //test.servicekey.key-3, v = 3
    err = configController.UpdateService("test.servicekey.key-3", 1, []string{}, []string{})
    if err == nil {
        t.Error("update on wrong version but success")
    }
}

func TestConfigController_RemoveService(t *testing.T) {
    //1：删除成功
    configController.RemoveService("test.servicekey.key-20", 21)
    if configController.IsServiceExist("test.servicekey.key-20") {
        t.Error("service test.servicekey.key-20 shouldn't exist, but exist")
    }
    num := len(configController.GetAllServiceKeys())
    if num != 99 {
        t.Errorf("after delete, expect 99 actual %d\n", num)
    }

    freeIndex := configController.GetFirstIndexFree()
    if freeIndex != 19 {
        t.Errorf("first free index expect 19, actual %d\n", freeIndex)
    }

    index, err := configController.QueryIndex("test.servicekey.key-20")
    if err == nil {
        t.Errorf("query test.servicekey.key-20's index shouldn't exist, but get %d\n", index)
    }

    //2：版本不对，删除失败
    configController.RemoveService("test.servicekey.key-3", 3)
    if !configController.IsServiceExist("test.servicekey.key-3") {
        t.Error("service test.servicekey.key-3 should exist, but not")
    }
    num = len(configController.GetAllServiceKeys())
    if num != 99 {
        t.Errorf("after delete, expect 99 actual %d\n", num)
    }

    //3：不存在，删除失败
    configController.RemoveService("adsadsad.adsada.sas", 1)
    num = len(configController.GetAllServiceKeys())
    if num != 99 {
        t.Errorf("after delete, expect 99 actual %d\n", num)
    }
}

func TestConfigController_ForceRemoveService(t *testing.T) {
    //1：删除成功
    configController.ForceRemoveService("test.servicekey.key-50")
    if configController.IsServiceExist("test.servicekey.key-50") {
        t.Error("service test.servicekey.key-50 shouldn't exist, but exist")
    }
    num := len(configController.GetAllServiceKeys())
    if num != 98 {
        t.Errorf("after delete, expect 98 actual %d\n", num)
    }
    configController.ForceRemoveService("xxxxxxx.xxxx.xx")
    num = len(configController.GetAllServiceKeys())
    if num != 98 {
        t.Errorf("after delete, expect 98 actual %d\n", num)
    }

    freeIndex := configController.GetFirstIndexFree()
    if freeIndex != 19 {
        t.Errorf("first free index expect 19, actual %d\n", freeIndex)
    }

    index, err := configController.QueryIndex("test.servicekey.key-50")
    if err == nil {
        t.Errorf("query test.servicekey.key-50's index shouldn't exist, but get %d\n", index)
    }
}

func TestMixed(t *testing.T) {
    originNum := len(configController.GetAllServiceKeys())
    //增2个
    configController.UpdateService("test.otherservicekey.key", 5, []string{"aaa.bbb","ccc.ddd"}, []string{"eee","fff"})
    configController.UpdateService("test.otherservicekey.key2", 5, []string{"aaa.bbb","ccc.ddd"}, []string{"eee","fff"})

    //减1个
    configController.ForceRemoveService("test.servicekey.key-10")
    //增1个
    configController.UpdateService("test.otherservicekey.key3", 5, []string{"aaa.bbb","ccc.ddd"}, []string{"eee","fff"})
    //减
    configController.ForceRemoveService("test.servicekey.key-96")
    //减
    configController.ForceRemoveService("test.servicekey.key-53")

    num := len(configController.GetAllServiceKeys())
    if originNum != num {
        t.Errorf("before operating, length %d, after, length %d\n", originNum, num)
    }
    freeIndex := configController.GetFirstIndexFree()
    if freeIndex != 52 {
        t.Errorf("first free index expect 52, actual %d\n", freeIndex)
    }
}

func TestConfigControllerClean(t *testing.T) {
    allServiceKeys := configController.GetAllServiceKeys()
    configController.Close()

    configController2, _ := GetConfigController()

    for serviceKey, version := range allServiceKeys {
        configController2.RemoveService(serviceKey, version + 1)
    }

    num := len(configController2.GetAllServiceKeys())
    if num != 0 {
        t.Errorf("after clean, number is %d\n", num)
    }
}

func TestMain(m *testing.M) {
    configController, _ = GetConfigController()

    for i := 0;i < 150;i++ {
        valuePrefix += "."
    }

    //set 99 config
    for i := 0;i < 99; i++ {
        serviceKey := fmt.Sprintf("test.servicekey.key-%d", i + 1)
        version := uint(i + 1)
        //have i+1 configures
        confKeys := make([]string, 0)
        values := make([]string, 0)
        for j := 0;j < i + 1; j++ {
            confKeys = append(confKeys, fmt.Sprintf("service-%d.conf-%d", i, j))
            values = append(values, fmt.Sprintf("%s-configure-value-%d", valuePrefix, j))
        }
        configController.UpdateService(serviceKey, version, confKeys, values)
    }
    m.Run()
}

