#include "sona_api.h"
#include "conf_memory.h"
#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <fcntl.h>
#include <string.h>
#include <pthread.h>
#include <sstream>
#include <algorithm>
#include "network.h"

static std::string trim(const std::string& str) {
    std::string::size_type pos = str.find_first_not_of(' ');
    if (pos == std::string::npos)
    {
        return "";
    }
    std::string::size_type pos2 = str.find_last_not_of(' ');
    if (pos2 != std::string::npos)
    {
        return str.substr(pos, pos2 - pos + 1);
    }
    return str.substr(pos);
}

static std::vector<std::string> split_by_comma(const std::string& str) {
    std::stringstream ss(str);
    std::vector<std::string> result;

    while(ss.good())
    {
        std::string substr;
        std::getline(ss, substr, ',');
        substr = trim(substr);
        if (substr != "") {
            result.push_back(substr);
        }
    }
    return result;
}

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
const std::string sona_api::get(std::string section, std::string key) {
    std::string conf_key = trim(section) + "." + trim(key);
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

//get list
std::vector<std::string> sona_api::get_list(std::string section, std::string key) {
    return split_by_comma(get(section, key));
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
