package org.sona.sys;

import java.io.File;
import java.nio.channels.FileChannel;
import java.nio.channels.FileLock;
import java.io.RandomAccessFile;
import java.io.IOException;

public class FileReadLock {
    private FileChannel channel;
    private FileLock lock;

    public FileReadLock(String filePath) throws Exception {
        File file = new File(filePath);
        RandomAccessFile raf = new RandomAccessFile(file,"r");
        channel = raf.getChannel();
        lock = null;
    }

    public void Lock() {
        while (true) {
            try {
                lock = channel.lock(0L, Long.MAX_VALUE, true);
            } catch (IOException e) {
                continue;
            }
            break;
        }
    }

    public void Release() {
        if (lock == null) {
            return ;
        }
        while (true) {
            try {
                lock.release();
            } catch (IOException e) {
                continue;
            }
            break;
        }
    }

    public static void main(String[] args) throws Exception {
        FileReadLock lock = new FileReadLock("/tmp/fuck.txt");
        System.out.println("try to get read lock");
        lock.Lock();
        System.out.println("get read lock ok");
        Thread.sleep(10000);
        System.out.println("release read lock");
        lock.Release();
    }
}

