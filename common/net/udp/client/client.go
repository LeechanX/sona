package client

import (
    "net"
    "fmt"
    "log"
    "time"
    "sona/common/net/protocol"
    "github.com/golang/protobuf/proto"
    "errors"
)

//比较简单，纯同步，支持俩接口：读、写
type Client struct {
    conn *net.UDPConn
}

func CreateClient(ip string, port int) (*Client, error) {
    //create UDP client
    addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
    if err != nil {
        return nil, err
    }
    conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        log.Panicf("dial udp error: %s\n", err)
        return nil, err
    }
    return &Client{conn:conn,}, nil
}

//收消息，指定消息头、消息PB类型
func (c *Client) Read(timeout time.Duration, cmdId uint, packet proto.Message) error {
    //接收 设置超时
    c.conn.SetReadDeadline(time.Now().Add(timeout))
    realCmdId, _, pbData, err := protocol.DecodeUDPMessage(c.conn)
    if err != nil {
        return err
    }
    if cmdId != realCmdId {
        return errors.New("udp receive error data")
    }
    //收到包
    if err = proto.Unmarshal(pbData, packet);err != nil {
        return err
    }
    return nil
}


//发消息
func (c *Client) Send(cmdId uint, pb proto.Message) error {
    data := protocol.EncodeMessage(cmdId, pb)
    c.conn.SetWriteDeadline(time.Now().Add(50 * time.Millisecond))
    length, error := c.conn.Write(data)
    if length < len(data) {
        return errors.New("write incompletely")
    }
    return error
}

func (c *Client) Close() {
    c.conn.Close()
}
