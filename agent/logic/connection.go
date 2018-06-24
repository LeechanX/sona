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
	CONNECTED = iota
	DISCONNECTED
)

type Connection struct {
	conn *net.TCPConn
	status uint32
	Wg sync.WaitGroup
	sendQueue chan protocol.PullConfigReq
	mutex sync.Mutex//用于保护CloseConnect操作，防止同时多个G进行CloseConnect
}

//创建一个connect结构体
func CreateConnect() *Connection {
	return &Connection{
		conn:nil,
		status:DISCONNECTED,
		sendQueue:nil,
	}
}

//执行连接
func (c *Connection) ConnectToBroker(addrStr string) error {
	if atomic.LoadUint32(&c.status) == CONNECTED {
		return errors.New("already connected with broker")
	}
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addrStr)
	log.Printf("connecting to broker %s\n", addrStr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalf("can's connect tcp address %s\n", addrStr)
		return err
	}
	log.Printf("connected to broker %s\n", addrStr)
	c.conn = conn
	//设置状态为已连接
	c.sendQueue = make(chan protocol.PullConfigReq, 1000)
	atomic.StoreUint32(&c.status, CONNECTED)
	return nil
}

//关闭连接
func (c *Connection) CloseConnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if atomic.LoadUint32(&c.status) == DISCONNECTED {
		log.Println("already closed the connection with broker")
		return
	}
	log.Println("now close the connection with broker")
	//设置连接已关闭
	atomic.StoreUint32(&c.status, DISCONNECTED)
	//关闭管道
	close(c.sendQueue)
	//关闭连接
	c.conn.Close()
}
