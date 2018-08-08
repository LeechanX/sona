import os
import sys
import fcntl
import struct

class FileReadlock(object):
    def __init__(self, index):
        start_len = 'qq'
        try:
            os.O_LARGEFILE
        except AttributeError:
            start_len = 'll'
        try:
            self.fd = os.open("/tmp/sona/cfg_{}.lock".format(index), os.O_RDONLY, 0o400)
        except Exception as e:
            raise e

        if sys.platform in ('netbsd1', 'netbsd2', 'netbsd3', 'Darwin1.2', 'darwin',
                            'freebsd2', 'freebsd3', 'freebsd4', 'freebsd5','freebsd6',
                            'freebsd7','bsdos2', 'bsdos3', 'bsdos4','openbsd', 'openbsd2',
                            'openbsd3'):
            if struct.calcsize('l') == 8:
                off_t = 'l'
                pid_t = 'i'
            else:
                off_t = 'lxxxx'
                pid_t = 'l'
            format = "%s%s%shh" % (off_t, off_t, pid_t)
            self.lock_op = struct.pack(format, 0, 0, 0, fcntl.F_RDLCK, 0)
            self.unlock_op = struct.pack(format, 0, 0, 0, fcntl.F_UNLCK, 0)
        else:
            format = "hh%shh" % start_len
            self.lock_op = struct.pack(format, fcntl.F_RDLCK, 0, 0, 0, 0, 0)
            self.unlock_op = struct.pack(format, fcntl.F_UNLCK, 0, 0, 0, 0, 0)

    '''
    lock the read lock
    '''
    def Lock(self):
        fcntl.fcntl(self.fd, fcntl.F_SETLKW, self.lock_op)
    '''
    unlock the read lock
    '''
    def Release(self):
        fcntl.fcntl(self.fd, fcntl.F_SETLK, self.unlock_op)

"""
try:
    lock = FileReadlock(3)
except Exception as e:
    print("wocao?", e)
print("try to lock...")
lock.Lock()
print("lock ok")

time.sleep(10)

print("release lock")
lock.Release()
"""