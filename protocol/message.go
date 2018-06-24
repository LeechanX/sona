package protocol

import (
	"log"
	"bytes"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
)

const HeadBytes = 8
const TotalLengthLimit = 102400

type MsgHead struct {
	CmdId MsgTypeId
	Length uint32
}

func EncodeMessage(cmdId MsgTypeId, pb proto.Message) []byte {
	//create head
	head := MsgHead{
		CmdId: cmdId,
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

func DecodeMessage(data []byte) (MsgTypeId, []byte, error) {
	buf := bytes.NewBuffer(data)
	//先读取包头
	head := MsgHead{}
	err := binary.Read(buf, binary.BigEndian, &head)
	if err != nil {
		log.Panicf("receive from udp data format error: %s\n", err)
		return MsgTypeId(0), nil, err
	}
	if head.Length <= HeadBytes || head.Length > TotalLengthLimit {
		log.Panicf("receive from udp data format error, length %d\n", head.Length)
		return MsgTypeId(0), nil, err
	}
	return head.CmdId, data[HeadBytes:], nil
}