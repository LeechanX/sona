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
