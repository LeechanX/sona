#ifndef __SONA_API_H__
#define __SONA_API_H__

#include "conf_memory.h"
#include <string>
#include <vector>

class sona_api {
public:
    ~sona_api();
    //get value
    const std::string get(std::string section, std::string key);
    std::vector<std::string> get_list(std::string section, std::string key);

    friend sona_api* init_api(const char* service_key);
    friend void* keep_using(void* args);
private:
    sona_api() {}

    int ffd;
    int sockfd;
    char *memory;
    unsigned index;
    std::string service_key;
};

sona_api* init_api(const char* service_key);

#endif
