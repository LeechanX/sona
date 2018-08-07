#include "sona_api.h"
#include "conf_memory.h"
#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <fcntl.h>
#include <string.h>
#include <pthread.h>
#include <algorithm>
#include "network.h"

//keep using
void* keep_using(void* args) {
    sona_api* api = (sona_api*)args;
    //keep using
    while (1)
    {
        sleep(10);
        //keep using
        send_keep_using(api->sockfd, api->service_key);
    }
}

sona_api::~sona_api() {
    if (ffd != -1) {
        close(ffd);
    }
    if (memory == NULL) {
        detach_mmap((void*)memory);
    }
    if (sockfd != -1) {
        close(sockfd);
    }
}

//get value
const std::string sona_api::api_get(std::string section, std::string key) {
    std::remove(section.begin(), section.end(), ' ');
    std::remove(key.begin(), key.end(), ' ');
    std::string conf_key = section + "." + key;
    if (conf_key.size() > ConfKeyCap) {
        fprintf(stderr, "format error: configure key %s is too long\n", conf_key.c_str());
        return "";
    }
    char c_value[ConfValueCap + 1];
    c_value[0] = '\0';
    //lock read lock
    struct flock lock;
    lock.l_type = F_RDLCK;
    lock.l_start = 0;
    lock.l_whence = SEEK_SET;
    lock.l_len = 0;
    fcntl(ffd, F_SETLKW, &lock);
    get_conf(memory, index, service_key.c_str(), conf_key.c_str(), c_value);
    //unlock
    lock.l_type = F_UNLCK;
    fcntl(ffd, F_SETLK, &lock);
    const std::string value(c_value);
    return value;
}

//must invoke init function before use api
sona_api* init_api(const char* service_key) {
    if (strlen(service_key) > ServiceKeyCap)
    {
        fprintf(stderr, "format error: service key %s is too long\n", service_key);
        return NULL;
    }
    sona_api* api = new sona_api();
    if (api == NULL) {
        return NULL;
    }
    api->sockfd = create_socket();
    if (api->sockfd == -1) {
        delete api;
        return NULL;
    }
    api->memory = (char*)attach_mmap();
    if (api->memory == NULL) {
        delete api;
        return NULL;
    }
    //subscribe
    int ret = subscribe(api->sockfd, service_key);
    if (ret == -1) {
        delete api;
        return NULL;
    }
    api->index = ret;
    //create flock
    char path[1024];
    sprintf(path, "/tmp/sona/cfg_%u.lock", api->index);
    api->ffd = open(path, O_RDONLY, 0400);
    if (api->ffd == -1)
    {
        delete api;
        return NULL;
    }
    api->service_key = service_key;
    pthread_t tid;
    pthread_create(&tid, NULL, keep_using, api);
    pthread_detach(tid);
    return api;
}
