package client

import (
    "fmt"
    "net"
    "log"
    "time"
    "errors"
    "sona/common/net/protocol"
    "github.com/golang/protobuf/proto"
)

type SyncClient struct {
    Ip string
    Port int
    conn *net.TCPConn
}

//创建一个client结构体并连接
func CreateSyncClient(ip string, port int) (*SyncClient, error) {
    c := &SyncClient{
        Ip:ip,
        Port:port,
        conn:nil,
    }
    //创建TCP客户端去连接broker
    tcpAddr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.Ip, c.Port))
    log.Printf("connecting to %s\n", fmt.Sprintf("%s:%d", c.Ip, c.Port))
    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        log.Printf("can's connect tcp address %s\n", fmt.Sprintf("%s:%d", c.Ip, c.Port))
        return nil, err
    }
    log.Printf("connected to broker %s successfully\n", fmt.Sprintf("%s:%d", c.Ip, c.Port))
    c.conn = conn
    return c, nil
}

//收消息，指定消息头、消息PB类型
func (c *SyncClient) Read(timeout time.Duration, cmdId uint, packet proto.Message) error {
    //接收 设置超时
    c.conn.SetReadDeadline(time.Now().Add(timeout))
    realCmdId, pbData, err := protocol.DecodeTCPMessage(c.conn)
    if err != nil {
        return err
    }
    if cmdId != realCmdId {
        return errors.New("tcp client receive error data")
    }
    //收到包
    if err = proto.Unmarshal(pbData, packet);err != nil {
        return err
    }
    return nil
}

//发送消息
func (c *SyncClient) Send(cmdId uint, pb proto.Message) error {
    data := protocol.EncodeMessage(cmdId, pb)
    c.conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
    length, err := c.conn.Write(data)
    if err != nil {
        return err
    }
    if length < len(data) {
        return errors.New("write incompletely")
    }
    return err
}

//关闭连接
func (c *SyncClient) Close() {
    //关闭连接
    c.conn.Close()
}
