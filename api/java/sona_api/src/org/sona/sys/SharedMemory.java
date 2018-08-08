package org.sona.sys;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.channels.FileChannel;
import java.nio.file.Paths;
import java.nio.file.StandardOpenOption;
import java.nio.MappedByteBuffer;

public class SharedMemory {
    private MappedByteBuffer mmap;
    private int index;
    public final int ServiceKeyCap = 92;
    //"section.key"组成服务的一个配置，各字段限长30字节
    private final int ConfKeyCap = 61;
    //value支持最大200字节
    private final int ConfValueCap = 200;
    private final int ServiceConfLimit = 100;
    //一个bucket用于存放一个Service的配置，可存多少种Service
    //[版本号:serviceKey长度:serviceKey:配置个数:配置k1长度:配置k1:配置v1长度:配置v1:......]
    //[2:2:92:2:[2:61:2:30]:[2:61:2:30]:...]
    private final int OneBucketCap = 2 + 2 + ServiceKeyCap + 2 + ServiceConfLimit * (2 + ConfKeyCap + 2 + ConfValueCap);

    public SharedMemory(int index) throws IOException {
        FileChannel channel = FileChannel.open(Paths.get("/tmp/sona/cfg.mmap"), StandardOpenOption.READ);
        //通过通道的map方法映射内存
        mmap = channel.map(FileChannel.MapMode.READ_ONLY, 0, channel.size());
        mmap.order(ByteOrder.BIG_ENDIAN);
        this.index = index;
    }

    /**
     * 获取某位置处的configure key
     * @param pos: 位置
     */
    private String GetConfKey(int pos) {
        int start = index * OneBucketCap + 6 + ServiceKeyCap + pos * (4 + ConfKeyCap + ConfValueCap);
        int len = mmap.getShort(start);
        start += 2;
        byte[] confKey = new byte[len];
        for (int i = 0;i < len; ++i) {
            confKey[i] = mmap.get(start + i);
        }
        return new String(confKey);
    }

    /**
     * 获取某位置处的configure value
     * @param pos: 位置
     */
    private String GetConfValue(int pos) {
        int start = index * OneBucketCap + 6 + ServiceKeyCap + pos * (4 + ConfKeyCap + ConfValueCap) + 2 + ConfKeyCap;
        int len = mmap.getShort(start);
        start += 2;
        byte[] value = new byte[len];
        for (int i = 0;i < len; ++i) {
            value[i] = mmap.get(start + i);
        }
        return new String(value);
    }

    /**
     * 二分搜索某key的位置
     * @param confKey: 配置key
     * */
    private int SearchOneConf(String confKey) {
        int start = index * OneBucketCap + 4 + ServiceKeyCap;
        int count = mmap.getShort(start);
        int low = 0, high = count;
        while (low < high) {
            int mid = (low + high) / 2;
            String key = GetConfKey(mid);
            int cmp = key.compareTo(confKey);
            if (cmp > 0) {
                high = mid;
            } else if (cmp < 0) {
                low = mid + 1;
            } else {
                return mid;
            }
        }
        return ServiceConfLimit;
    }

    /**
     * 判断是否存在服务配置
     * */
    private boolean HasService() {
        int start = index * OneBucketCap;
        byte[] s = new byte[2];
        for (int i = 0;i < 2;++i) {
            s[i] = mmap.get(i + start);
        }
        int version = mmap.getShort(start);
        return version != 0;
    }

    /***
     * 获取service key
     * */
    private String GetServiceKey() {
        int start = index * OneBucketCap + 2;
        mmap.order(ByteOrder.BIG_ENDIAN);
        int len = mmap.getShort(start);
        start += 2;
        byte[] serviceKey = new byte[len];
        for (int i = 0;i < len; ++i) {
            serviceKey[i] = mmap.get(i + start);
        }
        return new String(serviceKey);
    }

    /**
     * 获取value
     * @param serviceKey
     * @param confKey
     * */
    public String GetConf(String serviceKey, String confKey) {
        if (!HasService() || !serviceKey.equals(GetServiceKey())) {
            return "";
        }
        int pos = SearchOneConf(confKey);
        if (pos == ServiceConfLimit) {
            return "";
        }

        return GetConfValue(pos);
    }
}

