#include "sona_api.h"
#include <iostream>
#include <unistd.h>

int main() {
    sona_api* api = init_api("miliao.milink.pushgateway");
    if (api == NULL)
    {
        std::cout << "?" << std::endl;
        return 0;
    }
    for (int i = 0;i < 100; ++i) {
        std::string x = api->api_get("log", "level");
        std::cout << "log.level " << x << std::endl;
        x = api->api_get("pushstatus", "ip");
        std::cout << "pushstatus ip " << x << std::endl;
        x = api->api_get("pushstatus", "port");
        std::cout << "pushstatus port " << x << std::endl;
        sleep(1);
    }
    return 0;
}
