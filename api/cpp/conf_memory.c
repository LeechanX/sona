#include <stdio.h>
#include <fcntl.h>
#include <string.h>
#include <sys/mman.h>
#include <arpa/inet.h>
#include "conf_memory.h"

static unsigned short get_number(const char* slice) {
    unsigned short number = 0;
    memcpy(&number, slice, 2);
    return ntohs(number);
}

//return value is key, key's origin length is (ConfKeyCap + 1)
static void get_conf_key(const char* memory, unsigned index, unsigned pos, char* key) {
    unsigned start = OneBucketCap * index + 6 + ServiceKeyCap + pos * (4 + ConfKeyCap + ConfValueCap);
    unsigned short key_len = get_number(memory + start);//[start,start+2]
    strncpy(key, memory + start + 2, key_len);
    key[key_len] = '\0';
}

//return value is value, value's origin length is (ConfValueCap + 1)
static void get_conf_value(const char* memory, unsigned index, unsigned pos, char* value) {
    unsigned start = OneBucketCap * index + 6 + ServiceKeyCap + pos * ((4 + ConfKeyCap + ConfValueCap));
    //skip key
    start += 2 + ConfKeyCap;
    unsigned short value_len = get_number(memory + start);//start:start+2
    strncpy(value, memory + start + 2, value_len);
    value[value_len] = '\0';
}

//在某service下的配置中，找到target_key
static unsigned search_one_conf(const char* memory, unsigned index, const char* target_key) {
    unsigned start = OneBucketCap * index + 2 + 2 + ServiceKeyCap;
    unsigned offset;
    unsigned short conf_nums = get_number(memory + start);
    start += 2;
    unsigned short low = 0, high = conf_nums;
    unsigned short mid;

    char key[ConfKeyCap + 1];
    int cmp;

    while (low < high) {
        mid = (low + high) >> 1;
        //get key of mid
        get_conf_key(memory, index, mid, key);
        cmp = strcmp(target_key, key);
        if (cmp > 0) {
            low = mid + 1;
        } else if (cmp < 0) {
            high = mid;
        } else {
            return mid;//equal
        }
    }
    return ServiceConfLimit;
}

//某位置是否有配置
static int has_service(const char* memory, unsigned index) {
    unsigned start = OneBucketCap * index;
    //get version ?= 0
    unsigned short version = get_number(memory + start);
    return version != 0;
}

//get service key, return value is service_key (length is ServiceKeyCap + 1)
static void get_service_key(const char* memory, unsigned index, char* service_key) {
    unsigned start = OneBucketCap * index + 2;
    unsigned short service_key_len = get_number(memory + start);
    start += 2;
    strncpy(service_key, memory + start, service_key_len);
    service_key[service_key_len] = '\0';
}

//return value: 0 exist, -1 not
int get_conf(const char* memory, unsigned index,
        const char* service_key,
        const char* key,
        char* value) {
    char sk[ServiceKeyCap + 1];
    unsigned pos;
    value[0] = '\0';
    if (has_service(memory, index) == 0) {
        return -1;
    }
    get_service_key(memory, index, sk);
    if (strcmp(sk, service_key) != 0) {
        //is not service key
        return -1;
    }
    //binary search key
    pos = search_one_conf(memory, index, key);
    if (pos == ServiceConfLimit)
    {
        return -1;
    }
    get_conf_value(memory, index, pos, value);
    return 0;
}

void* attach_mmap() {
    //open as readonly
    int fd = open("/tmp/sona/cfg.mmap", O_RDONLY, 0400);
    if (fd == -1)
    {
        perror("open cfg.mmap");
        return NULL;
    }

    void* m = mmap(NULL, TotalConfMemSize, PROT_READ, MAP_SHARED, fd, 0);
    if (m == NULL)
    {
        perror("call mmap");
        return NULL;
    }
}

void detach_mmap(void* m) {
    munmap(m, TotalConfMemSize);
}
