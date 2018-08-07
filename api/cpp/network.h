#ifndef __NETWORK_H__
#define __NETWORK_H__

#include <string>

int create_socket();
void send_keep_using(int sockfd, const std::string& service_key);
int subscribe(int sockfd, const std::string& service_key);

#endif
