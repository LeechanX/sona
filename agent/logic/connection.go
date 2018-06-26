package logic

import (
	"net"
	"log"
	"sync"
	"errors"
	"sync/atomic"
	"easyconfig/protocol"
)

const (
	kConnStatusConnected = iota
	kConnStatusDisconnected
)

type Connection struct {
	conn *net.TCPConn
	status int32
	Wg sync.WaitGroup
	sendQueue chan *protocol.PullConfigReq
}

//创建一个connect结构体
func CreateConnect() *Connection {
	return &Connection{
		conn:nil,
		status:kConnStatusDisconnected,
		sendQueue:make(chan *protocol.PullConfigReq, 1000),
	}
}

//执行连接
func (c *Connection) ConnectToBroker(addrStr string) error {
	if atomic.LoadInt32(&c.status) == kConnStatusConnected {
		return errors.New("already connected with broker")
	}
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addrStr)
	log.Printf("connecting to broker %s\n", addrStr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Printf("can's connect tcp address %s\n", addrStr)
		return err
	}
	log.Printf("connected to broker %s\n", addrStr)
	c.conn = conn
	//设置状态为已连接
	atomic.StoreInt32(&c.status, kConnStatusConnected)
	return nil
}

//关闭连接
func (c *Connection) CloseConnect() {
	if !atomic.CompareAndSwapInt32(&c.status, kConnStatusConnected, kConnStatusDisconnected) {
		log.Println("already closed the connection with broker")
		return
	}
	log.Println("now close the connection with broker")
	//防止panic，不关闭管道，而是发送消息nil告知已无消息
	c.sendQueue<- nil
	//关闭连接
	c.conn.Close()
}
