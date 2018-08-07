#include <stdio.h>
#include <strings.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <stdint.h>
#include "base_protocol.pb.h"
#include "network.h"

struct MsgHead {
    uint32_t CmdId;
    uint32_t Length;
};

const uint32_t KeepUsingReqId = 1;
const uint32_t SubscribeReqId = 2;
const uint32_t SubscribeAgentRspId = 4;

int create_socket() {
    int fd = socket(AF_INET, SOCK_DGRAM, 0);
    if (fd == -1)
    {
        perror("create socket");
        return -1;
    }
    //call connect
    struct sockaddr_in servaddr;
    bzero(&servaddr, sizeof (servaddr));
    servaddr.sin_family = AF_INET;
    servaddr.sin_port = htons(9901);
    inet_aton("127.0.0.1", &servaddr.sin_addr);
    int ret = connect(fd, (struct sockaddr *)&servaddr, sizeof(servaddr));
    if (ret == -1)
    {
        perror("connect");
        return -1;
    }
    return fd;
}

void send_keep_using(int sockfd, const std::string& service_key) {
    protocol::KeepUsingReq req;
    req.set_servicekey(service_key);
    std::string req_str;
    req.SerializeToString(&req_str);
    MsgHead head;
    head.CmdId = htonl(KeepUsingReqId);
    unsigned length = 8 + req_str.size();
    head.Length = htonl(length);
    char send_buf[1024];
    memcpy(send_buf, &head, 8);
    memcpy(send_buf + 8, req_str.c_str(), req_str.size());
    sendto(sockfd, send_buf, length, 0, NULL, 0);
}

int subscribe(int sockfd, const std::string& service_key) {
    protocol::SubscribeReq req;
    req.set_servicekey(service_key);
    std::string req_str;
    req.SerializeToString(&req_str);
    MsgHead head;
    head.CmdId = htonl(SubscribeReqId);
    unsigned length = 8 + req_str.size();
    head.Length = htonl(length);

    char send_buf[1024];
    memcpy(send_buf, &head, 8);
    memcpy(send_buf + 8, req_str.c_str(), req_str.size());
    int sn = sendto(sockfd, send_buf, length, 0, NULL, 0);
    if (sn != length) {
        return -1;
    }
    struct timeval timo;
    timo.tv_sec = 0;
    timo.tv_usec = 300000;//300ms
    setsockopt(sockfd, SOL_SOCKET, SO_RCVTIMEO, &timo, sizeof(timo));
    char recv_buf[1024];
    int rn = recvfrom(sockfd, recv_buf, 1024, 0, NULL, NULL);
    if (rn == -1)
    {
        perror("call recvfrom");
        return -1;
    }
    if (rn <= 8) {
        fprintf(stderr, "data is not complete\n");
        return -1;
    }
    memcpy(&head, recv_buf, 8);
    if (ntohl(head.CmdId) != SubscribeAgentRspId) {
        fprintf(stderr, "cmdid is not corret\n");
        return -1;
    }
    int data_len = ntohl(head.Length) - 8;
    protocol::SubscribeAgentRsp rsp;
    rsp.ParseFromArray(recv_buf + 8, data_len);
    if (rsp.code() == 0)
    {
        //订阅成功, get index
        return rsp.index();
    }
    fprintf(stderr, "subscribe error: no data\n");
    return -1;
}
