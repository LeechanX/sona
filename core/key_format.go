package core

import "strings"

func IsValidityServiceKey(key string) bool {
    //prductionname.groupname.servicename
    s := strings.Split(key, ".")
    if len(s) != 3 {
        return false
    }
    for _, field := range s {
        if field == "" {
            return false
        }
    }
    return true
}

func IsValidityConfKey(key string) bool {
    //section.key
    s := strings.Split(key, ".")
    if len(s) != 2 {
        return false
    }
    for _, field := range s {
        if field == "" {
            return false
        }
    }
    return true
}


//splice total key
func SpliceKey(serviceKey string, configKey string) string {
    return serviceKey + "." + configKey
}

//extract the service key
func GetServiceKey(key string) string {
    s := strings.Split(key, ".")
    return strings.Join(s[:3], ".")
}

//获取服务内的配置名
func GetConfigKey(key string) string {
    s := strings.Split(key, ".")
    return strings.Join(s[3:], ".")
}