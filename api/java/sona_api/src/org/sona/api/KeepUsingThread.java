package org.sona.api;

import org.sona.net.UDPClient;

public class KeepUsingThread extends Thread {
    private UDPClient client;
    KeepUsingThread(UDPClient c) {
        this.client = c;
    }

    public void run() {
        while (true) {
            try {
                Thread.sleep(1000 * 60);
                client.KeepUsing();
            } catch (InterruptedException e) {
            }
        }
    }
}

