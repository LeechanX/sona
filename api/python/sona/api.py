#encoding=utf-8

import time
import threading

import udpclient
import memory
import flock

'''
保活线程
'''
class BackendThread(threading.Thread):
    def __init__(self, u):
        threading.Thread.__init__(self)
        self.u = u
    def run(self):
        while True:
            time.sleep(1)
            print("fuck")
            self.u.keep_using()

class SonaApi:
    def __init__(self, service_key):
        self.service_key = service_key
        #udp client
        self.u = udpclient.UDPClient(service_key)
        #subscribe service
        index = self.u.subscribe()
        if index == -1:
            raise Exception("service key is not exist")
        #attach shared memory
        self.mmap = memory.SharedMemory(service_key, index)
        #create read flock
        self.read_lock = flock.FileReadlock(index)
        #backup thread
        backend = BackendThread(self.u)
        backend.setDaemon(True)
        backend.start()
    '''
    get configure value
    '''
    def get(self, section, key):
        conf_key = section.strip() + "." + key.strip()
        if len(conf_key) > memory.ConfKeyCap:
            return ""
        self.read_lock.Lock()
        value = self.mmap.get_conf(conf_key)
        self.read_lock.Release()
        return value
    '''
    get configure value list
    '''
    def get_list(self, section, key):
        value = self.get(section, key)
        items = value.split()
        return [item.strip() for item in items if item.strip() != ""]

'''
api = SonaApi("lebron.james.info")
print(api.get_list("friends", "list"))
print(api.get("player", "team"))
'''
