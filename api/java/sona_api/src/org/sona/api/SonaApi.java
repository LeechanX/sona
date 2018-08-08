package org.sona.api;

import org.sona.net.UDPClient;
import org.sona.sys.SharedMemory;
import org.sona.sys.FileReadLock;

import java.util.ArrayList;

public class SonaApi {
    private String serviceKey;
    private SharedMemory memory;
    private FileReadLock lock;

    public SonaApi(String serviceKey) throws Exception {
        this.serviceKey = serviceKey;
        UDPClient client = new UDPClient(serviceKey);
        int index = client.Subscribe();

        if (index == -1) {
            throw new Exception("subscribe service key failed");
        }

        //attach shared memory
        memory = new SharedMemory(index);
        //create file lock
        String flockPath = "/tmp/sona/cfg_" + index + ".lock";
        lock = new FileReadLock(flockPath);

        KeepUsingThread thread = new KeepUsingThread(client);
        thread.start();
    }

    public String Get(String section, String key) {
        String confKey = section.trim() + "." + key.trim();
        if (confKey.length() > memory.ServiceKeyCap) {
            return "";
        }
        //lock
        lock.Lock();
        String value = memory.GetConf(serviceKey, confKey);
        //release
        lock.Release();
        return value;
    }

    public ArrayList<String> GetList(String section, String key) {
        String value = Get(section, key);
        ArrayList<String> list = new ArrayList<>();
        String[] items = value.split(",");
        for (String item: items) {
            if (!item.trim().equals("")) {
                list.add(item);
            }
        }
        return list;
    }
}

