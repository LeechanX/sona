package common

import "sort"

type ConfigureItem struct {
    key string
    value string
}

type ConfigList []*ConfigureItem

func (cl ConfigList) Len() int {
    return len(cl)
}

func (cl ConfigList) Less(i, j int) bool {
    return cl[i].key < cl[j].key
}

func (cl ConfigList) Swap(i, j int) {
    cl[i], cl[j] = cl[j], cl[i]
}

func SortKV(keys, values []string) ([]string, []string) {
    sortedKeys := make([]string, 0)
    sortedValues := make([]string, 0)

    sortedKV := make([]*ConfigureItem, 0)
    for i := 0;i < len(keys);i++ {
        sortedKV = append(sortedKV, &ConfigureItem{key:keys[i], value:values[i]})
    }
    sort.Sort(ConfigList(sortedKV))
    for _, kv := range sortedKV {
        sortedKeys = append(sortedKeys, kv.key)
        sortedValues = append(sortedValues, kv.value)
    }
    return sortedKeys, sortedValues
}
