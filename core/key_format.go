package core

import "strings"

func IsValidityKey(key string) bool {
    //prductionname.groupname.servicename.section.key
    s := strings.Split(key, ".")
    if len(s) != FieldNumber {
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