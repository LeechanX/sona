import sys
import struct
import socket
import base_protocol_pb2

class UDPClient(object):
    def __init__(self, service_key):
        self.s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.service_key = service_key
    '''
    send keep using request
    '''
    def keep_using(self):
        KeepUsingReqId = 1
        req = base_protocol_pb2.KeepUsingReq()
        req.serviceKey = self.service_key
        body = req.SerializeToString()
        cmdid_str = struct.pack('i', socket.htonl(KeepUsingReqId))
        length_str = struct.pack('i', socket.htonl(8 + len(body)))
        try:
            self.s.sendto(cmdid_str + length_str + body, ('127.0.0.1', 9901))
        except socket.error as e:
            print("", e)
            #print(e, file = sys.stderr)
    '''
    send subscribe request and receive result
    '''
    def subscribe(self):
        SubscribeReqId = 2
        SubscribeAgentRspId = 4
        #send request
        req = base_protocol_pb2.SubscribeReq()
        req.serviceKey = self.service_key
        body = req.SerializeToString()
        cmdid_str = struct.pack('i', socket.htonl(SubscribeReqId))
        length_str = struct.pack('i', socket.htonl(8 + len(body)))
        try:
            self.s.sendto(cmdid_str + length_str + body, ('127.0.0.1', 9901))
        except socket.error as e:
            print(e)
            return -1
        #receive response 300ms timeout
        try:
            self.s.settimeout(0.3)
            rsp_str, _ = self.s.recvfrom(1024)
        except socket.timeout as e:
            print(e)
            return -1
        if len(rsp_str) <= 8:
            print('received uncomplete packet')
            return -1
        cmdid = socket.ntohl(struct.unpack('i', rsp_str[:4])[0])
        if cmdid != SubscribeAgentRspId:
            print('received unknown cmdid: {}'.format(cmdid))
            return -1
        length = socket.ntohl(struct.unpack('i', rsp_str[4:8])[0])
        if length != len(rsp_str):
            print('received unexpect length: {}'.format(length))
            return -1
        rsp = base_protocol_pb2.SubscribeAgentRsp()
        rsp.ParseFromString(rsp_str[8:])
        if rsp.code == 0:
            return rsp.index
        return -1

'''
client = UDPClient('miliao.milink.pushgateway')
print(client.keep_using())
print(client.subscribe())
'''