package core

import "strings"

func IsValidityKey(key string) bool {
    //prductionname.groupname.servicename.section.key
    s := strings.Split(key, ".")
    if len(s) != 5 {
        return false
    }
    for _, field := range(s) {
        if field == "" {
            return false
        }
    }
    return true
}

//splice total key
func GetTotalKey(serviceId string, section string, key string) string {
    return serviceId + "." + section + "." + key
}

//extract the service Id
func GetServiceId(key string) string {
    s := strings.Split(key, ".")
    return strings.Join(s[:3], ".")
}
