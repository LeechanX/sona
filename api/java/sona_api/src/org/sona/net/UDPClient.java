package org.sona.net;

import java.net.*;
import java.io.IOException;
import java.nio.ByteOrder;
import java.nio.ByteBuffer;
import org.sona.protocol.BaseProtocol;
import com.google.protobuf.InvalidProtocolBufferException;

public class UDPClient {
    private DatagramSocket socket;
    private InetAddress address;
    private String serviceKey;

    public UDPClient(String serviceKey) throws SocketException {
        this.serviceKey = serviceKey;
        socket = new DatagramSocket();
        try {
            address = InetAddress.getByName("127.0.0.1");
        } catch (UnknownHostException e) {

        }
        socket.connect(address, 9901);
    }

    /**
     * 向agent订阅某service key
     */
    public int Subscribe() {
        final int SubscribeReqId = 2;
        BaseProtocol.SubscribeReq.Builder builder = BaseProtocol.SubscribeReq.newBuilder();
        builder.setServiceKey(serviceKey);
        BaseProtocol.SubscribeReq req = builder.build();
        byte[] packet = MakePacket(SubscribeReqId, req.toByteArray());
        //发送订阅请求
        DatagramPacket sendPacket=new DatagramPacket(packet, packet.length, address, 9901);
        try {
            socket.send(sendPacket);
        } catch (IOException e) {
            e.printStackTrace();
            return -1;
        }

        final int SubscribeAgentRspId = 4;
        byte[] recvBuf=new byte[1024];
        DatagramPacket recvPacket = new DatagramPacket(recvBuf, recvBuf.length);

        //接收订阅结果
        //300ms超时
        try {
            socket.setSoTimeout(300);
            socket.receive(recvPacket);
        } catch (SocketTimeoutException e) {
            e.printStackTrace();
            return -1;
        } catch (SocketException e) {
            e.printStackTrace();
            return -1;
        } catch (IOException e) {
            e.printStackTrace();
            return -1;
        }
        BaseProtocol.SubscribeAgentRsp rsp = null;
        byte[] response = recvPacket.getData();
        int responseLen = recvPacket.getLength();
        if (responseLen <= 8) {
            System.err.println("receive uncomplete packet");
            return -1;
        }

        ByteBuffer cmdField = ByteBuffer.wrap(response, 0, 4);
        cmdField.order(ByteOrder.BIG_ENDIAN);
        if (cmdField.asIntBuffer().get() != SubscribeAgentRspId) {
            System.err.printf("receive packet head cmdid is %d\n", cmdField.asIntBuffer().get());
            return -1;
        }

        ByteBuffer lengthField = ByteBuffer.wrap(response, 4, 8);
        lengthField.order(ByteOrder.BIG_ENDIAN);
        if (lengthField.asIntBuffer().get() != responseLen) {
            System.err.printf("receive packet length is %d\n", lengthField.asIntBuffer().get());
            return -1;
        }

        byte[] pb = new byte[responseLen - 8];
        System.arraycopy(response, 8, pb, 0, responseLen - 8);

        try {
            rsp = BaseProtocol.SubscribeAgentRsp.parseFrom(pb);
        } catch (InvalidProtocolBufferException e) {
            System.out.println("decode SubscribeAgentRsp error");
            return -1;
        }
        if (rsp.getCode() == 0) {
            //订阅成功
            return rsp.getIndex();
        }
        return -1;
    }

    /**
     * 向agent报告“在使用serviceKey”
     */
    public void KeepUsing() {
        final int KeepUsingReqId = 1;
        BaseProtocol.KeepUsingReq.Builder builder = BaseProtocol.KeepUsingReq.newBuilder();
        builder.setServiceKey(serviceKey);
        BaseProtocol.KeepUsingReq req = builder.build();
        byte[] packet = MakePacket(KeepUsingReqId, req.toByteArray());
        DatagramPacket sendPacket=new DatagramPacket(packet, packet.length, address, 9901);
        try {
            socket.send(sendPacket);
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    private byte[] MakePacket(final int cmdid, byte[] body) {
        //构造head
        ByteBuffer cmdField = ByteBuffer.allocate(4);
        cmdField.order(ByteOrder.BIG_ENDIAN);
        cmdField.asIntBuffer().put(cmdid);
        ByteBuffer lenField = ByteBuffer.allocate(4);
        lenField.order(ByteOrder.BIG_ENDIAN);
        lenField.asIntBuffer().put(body.length + 8);

        byte[] packet = new byte[cmdField.array().length + lenField.array().length + body.length];
        System.arraycopy(cmdField.array(), 0, packet, 0, cmdField.array().length);
        System.arraycopy(lenField.array(), 0, packet, cmdField.array().length, lenField.array().length);
        System.arraycopy(body, 0, packet, cmdField.array().length + lenField.array().length, body.length);
        return packet;
    }
}

