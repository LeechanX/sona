package protocol

import (
    "io"
    "net"
    "fmt"
    "time"
    "bytes"
    "errors"
    "encoding/binary"
    "github.com/golang/protobuf/proto"
)

const (
    HeadBytes = 8
    TotalLengthLimit = 102400
)

type MsgHead struct {
    CmdId uint32//必须是明确字节数的类型，不然buf.Read可能有问题
    Length uint32
}

func EncodeMessage(cmdId uint, pb proto.Message) []byte {
    //create head
    head := MsgHead{
        CmdId: uint32(cmdId),
        Length: HeadBytes,
    }
    //create body
    pbData, _ := proto.Marshal(pb)
    //update package length
    head.Length += uint32(len(pbData))
    //拼包
    buf := &bytes.Buffer{}
    binary.Write(buf, binary.BigEndian, &head)
    buf.Write(pbData)
    return buf.Bytes()
}

func DecodeUDPMessage(conn *net.UDPConn) (uint, *net.UDPAddr, []byte, error) {
    data := make([]byte, TotalLengthLimit)
    nBytes, cliAddr, err := conn.ReadFromUDP(data)
    if err != nil {
        return 0, nil, nil, err
    }

    if nBytes <= HeadBytes {
        return 0, nil, nil, errors.New(fmt.Sprintf(
            "receive from udp length error: %d\n", nBytes))
    }
    buf := bytes.NewBuffer(data)
    //先读取包头
    head := MsgHead{}
    err = binary.Read(buf, binary.BigEndian, &head)
    if err != nil {
        return 0, nil, nil, errors.New(fmt.Sprintf(
            "receive from udp data format error: %s\n", err))
    }
    if head.Length <= HeadBytes || head.Length > TotalLengthLimit {
        return 0, nil, nil, errors.New(fmt.Sprintf(
            "receive from udp data format error, length %d\n", head.Length))
    }
    return uint(head.CmdId), cliAddr, data[HeadBytes:head.Length], nil
}

func DecodeTCPMessage(conn *net.TCPConn) (uint, []byte, error) {
    data := make([]byte, TotalLengthLimit)
    //tcpConn.Read和io.ReadFull的区别，很关键
    //先读取包头
    _, err := io.ReadFull(conn, data[:HeadBytes])
    if err != nil {
        return 0, nil, err
    }
    buf := bytes.NewBuffer(data[:HeadBytes])
    head := MsgHead{}
    err = binary.Read(buf, binary.BigEndian, &head)
    if err != nil {
        return 0, nil, err
    }
    if head.Length <= HeadBytes || head.Length > TotalLengthLimit {
        return 0, nil, errors.New(fmt.Sprintf(
            "receive from tcp data format error, length %d\n", head.Length))
    }
    bodyLength := head.Length - HeadBytes

    //再读取body 这里最好设置50ms超时，因为包头已经来了，包身不应该太久不来
    conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
    _, err = io.ReadFull(conn, data[:bodyLength])
    //清除超时
    conn.SetReadDeadline(time.Time{})
    if err != nil {
        return 0, nil, err
    }
    return uint(head.CmdId), data[:bodyLength], nil
}
