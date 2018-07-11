// Code generated by protoc-gen-go. DO NOT EDIT.
// source: base_protocol.proto

package protocol

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type MsgTypeId int32

const (
	MsgTypeId_KeepUsingReqId         MsgTypeId = 1
	MsgTypeId_SubscribeReqId         MsgTypeId = 2
	MsgTypeId_SubscribeBrokerRspId   MsgTypeId = 3
	MsgTypeId_SubscribeAgentRspId    MsgTypeId = 4
	MsgTypeId_AddConfigReqId         MsgTypeId = 5
	MsgTypeId_DelConfigReqId         MsgTypeId = 6
	MsgTypeId_UpdateConfigReqId      MsgTypeId = 7
	MsgTypeId_PullServiceConfigReqId MsgTypeId = 8
	MsgTypeId_PullServiceConfigRspId MsgTypeId = 9
)

var MsgTypeId_name = map[int32]string{
	1: "KeepUsingReqId",
	2: "SubscribeReqId",
	3: "SubscribeBrokerRspId",
	4: "SubscribeAgentRspId",
	5: "AddConfigReqId",
	6: "DelConfigReqId",
	7: "UpdateConfigReqId",
	8: "PullServiceConfigReqId",
	9: "PullServiceConfigRspId",
}
var MsgTypeId_value = map[string]int32{
	"KeepUsingReqId":         1,
	"SubscribeReqId":         2,
	"SubscribeBrokerRspId":   3,
	"SubscribeAgentRspId":    4,
	"AddConfigReqId":         5,
	"DelConfigReqId":         6,
	"UpdateConfigReqId":      7,
	"PullServiceConfigReqId": 8,
	"PullServiceConfigRspId": 9,
}

func (x MsgTypeId) Enum() *MsgTypeId {
	p := new(MsgTypeId)
	*p = x
	return p
}
func (x MsgTypeId) String() string {
	return proto.EnumName(MsgTypeId_name, int32(x))
}
func (x *MsgTypeId) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(MsgTypeId_value, data, "MsgTypeId")
	if err != nil {
		return err
	}
	*x = MsgTypeId(value)
	return nil
}
func (MsgTypeId) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_base_protocol_013846fdbab5dc65, []int{0}
}

// client向agent发起使用中请求
type KeepUsingReq struct {
	ServiceKey           *string  `protobuf:"bytes,1,req,name=serviceKey" json:"serviceKey,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KeepUsingReq) Reset()         { *m = KeepUsingReq{} }
func (m *KeepUsingReq) String() string { return proto.CompactTextString(m) }
func (*KeepUsingReq) ProtoMessage()    {}
func (*KeepUsingReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_base_protocol_013846fdbab5dc65, []int{0}
}
func (m *KeepUsingReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeepUsingReq.Unmarshal(m, b)
}
func (m *KeepUsingReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeepUsingReq.Marshal(b, m, deterministic)
}
func (dst *KeepUsingReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeepUsingReq.Merge(dst, src)
}
func (m *KeepUsingReq) XXX_Size() int {
	return xxx_messageInfo_KeepUsingReq.Size(m)
}
func (m *KeepUsingReq) XXX_DiscardUnknown() {
	xxx_messageInfo_KeepUsingReq.DiscardUnknown(m)
}

var xxx_messageInfo_KeepUsingReq proto.InternalMessageInfo

func (m *KeepUsingReq) GetServiceKey() string {
	if m != nil && m.ServiceKey != nil {
		return *m.ServiceKey
	}
	return ""
}

// client向agent发起订阅请求,agent向broker发起订阅请求
type SubscribeReq struct {
	ServiceKey           *string  `protobuf:"bytes,1,req,name=serviceKey" json:"serviceKey,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubscribeReq) Reset()         { *m = SubscribeReq{} }
func (m *SubscribeReq) String() string { return proto.CompactTextString(m) }
func (*SubscribeReq) ProtoMessage()    {}
func (*SubscribeReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_base_protocol_013846fdbab5dc65, []int{1}
}
func (m *SubscribeReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscribeReq.Unmarshal(m, b)
}
func (m *SubscribeReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscribeReq.Marshal(b, m, deterministic)
}
func (dst *SubscribeReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscribeReq.Merge(dst, src)
}
func (m *SubscribeReq) XXX_Size() int {
	return xxx_messageInfo_SubscribeReq.Size(m)
}
func (m *SubscribeReq) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscribeReq.DiscardUnknown(m)
}

var xxx_messageInfo_SubscribeReq proto.InternalMessageInfo

func (m *SubscribeReq) GetServiceKey() string {
	if m != nil && m.ServiceKey != nil {
		return *m.ServiceKey
	}
	return ""
}

// broker向agent回复订阅请求
type SubscribeBrokerRsp struct {
	ServiceKey           *string  `protobuf:"bytes,1,req,name=serviceKey" json:"serviceKey,omitempty"`
	Code                 *int32   `protobuf:"varint,2,req,name=code" json:"code,omitempty"`
	Version              *uint32  `protobuf:"varint,3,req,name=version" json:"version,omitempty"`
	ConfKeys             []string `protobuf:"bytes,4,rep,name=confKeys" json:"confKeys,omitempty"`
	Values               []string `protobuf:"bytes,5,rep,name=values" json:"values,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubscribeBrokerRsp) Reset()         { *m = SubscribeBrokerRsp{} }
func (m *SubscribeBrokerRsp) String() string { return proto.CompactTextString(m) }
func (*SubscribeBrokerRsp) ProtoMessage()    {}
func (*SubscribeBrokerRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_base_protocol_013846fdbab5dc65, []int{2}
}
func (m *SubscribeBrokerRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscribeBrokerRsp.Unmarshal(m, b)
}
func (m *SubscribeBrokerRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscribeBrokerRsp.Marshal(b, m, deterministic)
}
func (dst *SubscribeBrokerRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscribeBrokerRsp.Merge(dst, src)
}
func (m *SubscribeBrokerRsp) XXX_Size() int {
	return xxx_messageInfo_SubscribeBrokerRsp.Size(m)
}
func (m *SubscribeBrokerRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscribeBrokerRsp.DiscardUnknown(m)
}

var xxx_messageInfo_SubscribeBrokerRsp proto.InternalMessageInfo

func (m *SubscribeBrokerRsp) GetServiceKey() string {
	if m != nil && m.ServiceKey != nil {
		return *m.ServiceKey
	}
	return ""
}

func (m *SubscribeBrokerRsp) GetCode() int32 {
	if m != nil && m.Code != nil {
		return *m.Code
	}
	return 0
}

func (m *SubscribeBrokerRsp) GetVersion() uint32 {
	if m != nil && m.Version != nil {
		return *m.Version
	}
	return 0
}

func (m *SubscribeBrokerRsp) GetConfKeys() []string {
	if m != nil {
		return m.ConfKeys
	}
	return nil
}

func (m *SubscribeBrokerRsp) GetValues() []string {
	if m != nil {
		return m.Values
	}
	return nil
}

// agent向client回复订阅请求
type SubscribeAgentRsp struct {
	ServiceKey           *string  `protobuf:"bytes,1,req,name=serviceKey" json:"serviceKey,omitempty"`
	Code                 *int32   `protobuf:"varint,2,req,name=code" json:"code,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubscribeAgentRsp) Reset()         { *m = SubscribeAgentRsp{} }
func (m *SubscribeAgentRsp) String() string { return proto.CompactTextString(m) }
func (*SubscribeAgentRsp) ProtoMessage()    {}
func (*SubscribeAgentRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_base_protocol_013846fdbab5dc65, []int{3}
}
func (m *SubscribeAgentRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscribeAgentRsp.Unmarshal(m, b)
}
func (m *SubscribeAgentRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscribeAgentRsp.Marshal(b, m, deterministic)
}
func (dst *SubscribeAgentRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscribeAgentRsp.Merge(dst, src)
}
func (m *SubscribeAgentRsp) XXX_Size() int {
	return xxx_messageInfo_SubscribeAgentRsp.Size(m)
}
func (m *SubscribeAgentRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscribeAgentRsp.DiscardUnknown(m)
}

var xxx_messageInfo_SubscribeAgentRsp proto.InternalMessageInfo

func (m *SubscribeAgentRsp) GetServiceKey() string {
	if m != nil && m.ServiceKey != nil {
		return *m.ServiceKey
	}
	return ""
}

func (m *SubscribeAgentRsp) GetCode() int32 {
	if m != nil && m.Code != nil {
		return *m.Code
	}
	return 0
}

// broker向client推送添加一个配置项命令
type AddConfigReq struct {
	ServiceKey           *string  `protobuf:"bytes,1,req,name=serviceKey" json:"serviceKey,omitempty"`
	Version              *uint32  `protobuf:"varint,2,req,name=version" json:"version,omitempty"`
	ConfKey              *string  `protobuf:"bytes,3,req,name=confKey" json:"confKey,omitempty"`
	Value                *string  `protobuf:"bytes,4,req,name=value" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AddConfigReq) Reset()         { *m = AddConfigReq{} }
func (m *AddConfigReq) String() string { return proto.CompactTextString(m) }
func (*AddConfigReq) ProtoMessage()    {}
func (*AddConfigReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_base_protocol_013846fdbab5dc65, []int{4}
}
func (m *AddConfigReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AddConfigReq.Unmarshal(m, b)
}
func (m *AddConfigReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AddConfigReq.Marshal(b, m, deterministic)
}
func (dst *AddConfigReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AddConfigReq.Merge(dst, src)
}
func (m *AddConfigReq) XXX_Size() int {
	return xxx_messageInfo_AddConfigReq.Size(m)
}
func (m *AddConfigReq) XXX_DiscardUnknown() {
	xxx_messageInfo_AddConfigReq.DiscardUnknown(m)
}

var xxx_messageInfo_AddConfigReq proto.InternalMessageInfo

func (m *AddConfigReq) GetServiceKey() string {
	if m != nil && m.ServiceKey != nil {
		return *m.ServiceKey
	}
	return ""
}

func (m *AddConfigReq) GetVersion() uint32 {
	if m != nil && m.Version != nil {
		return *m.Version
	}
	return 0
}

func (m *AddConfigReq) GetConfKey() string {
	if m != nil && m.ConfKey != nil {
		return *m.ConfKey
	}
	return ""
}

func (m *AddConfigReq) GetValue() string {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return ""
}

// broker向client发起删除一个配置项命令
type DelConfigReq struct {
	ServiceKey           *string  `protobuf:"bytes,1,req,name=serviceKey" json:"serviceKey,omitempty"`
	Version              *uint32  `protobuf:"varint,2,req,name=version" json:"version,omitempty"`
	ConfKey              *string  `protobuf:"bytes,3,req,name=confKey" json:"confKey,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DelConfigReq) Reset()         { *m = DelConfigReq{} }
func (m *DelConfigReq) String() string { return proto.CompactTextString(m) }
func (*DelConfigReq) ProtoMessage()    {}
func (*DelConfigReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_base_protocol_013846fdbab5dc65, []int{5}
}
func (m *DelConfigReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DelConfigReq.Unmarshal(m, b)
}
func (m *DelConfigReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DelConfigReq.Marshal(b, m, deterministic)
}
func (dst *DelConfigReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DelConfigReq.Merge(dst, src)
}
func (m *DelConfigReq) XXX_Size() int {
	return xxx_messageInfo_DelConfigReq.Size(m)
}
func (m *DelConfigReq) XXX_DiscardUnknown() {
	xxx_messageInfo_DelConfigReq.DiscardUnknown(m)
}

var xxx_messageInfo_DelConfigReq proto.InternalMessageInfo

func (m *DelConfigReq) GetServiceKey() string {
	if m != nil && m.ServiceKey != nil {
		return *m.ServiceKey
	}
	return ""
}

func (m *DelConfigReq) GetVersion() uint32 {
	if m != nil && m.Version != nil {
		return *m.Version
	}
	return 0
}

func (m *DelConfigReq) GetConfKey() string {
	if m != nil && m.ConfKey != nil {
		return *m.ConfKey
	}
	return ""
}

// broker向client发起更新一个配置项的命令
type UpdateConfigReq struct {
	ServiceKey           *string  `protobuf:"bytes,1,req,name=serviceKey" json:"serviceKey,omitempty"`
	Version              *uint32  `protobuf:"varint,2,req,name=version" json:"version,omitempty"`
	ConfKey              *string  `protobuf:"bytes,3,req,name=confKey" json:"confKey,omitempty"`
	Value                *string  `protobuf:"bytes,4,req,name=value" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UpdateConfigReq) Reset()         { *m = UpdateConfigReq{} }
func (m *UpdateConfigReq) String() string { return proto.CompactTextString(m) }
func (*UpdateConfigReq) ProtoMessage()    {}
func (*UpdateConfigReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_base_protocol_013846fdbab5dc65, []int{6}
}
func (m *UpdateConfigReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateConfigReq.Unmarshal(m, b)
}
func (m *UpdateConfigReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateConfigReq.Marshal(b, m, deterministic)
}
func (dst *UpdateConfigReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateConfigReq.Merge(dst, src)
}
func (m *UpdateConfigReq) XXX_Size() int {
	return xxx_messageInfo_UpdateConfigReq.Size(m)
}
func (m *UpdateConfigReq) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateConfigReq.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateConfigReq proto.InternalMessageInfo

func (m *UpdateConfigReq) GetServiceKey() string {
	if m != nil && m.ServiceKey != nil {
		return *m.ServiceKey
	}
	return ""
}

func (m *UpdateConfigReq) GetVersion() uint32 {
	if m != nil && m.Version != nil {
		return *m.Version
	}
	return 0
}

func (m *UpdateConfigReq) GetConfKey() string {
	if m != nil && m.ConfKey != nil {
		return *m.ConfKey
	}
	return ""
}

func (m *UpdateConfigReq) GetValue() string {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return ""
}

// agent向broker拉取一个服务的最新配置
type PullServiceConfigReq struct {
	ServiceKey           *string  `protobuf:"bytes,1,req,name=serviceKey" json:"serviceKey,omitempty"`
	Version              *uint32  `protobuf:"varint,2,req,name=version" json:"version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PullServiceConfigReq) Reset()         { *m = PullServiceConfigReq{} }
func (m *PullServiceConfigReq) String() string { return proto.CompactTextString(m) }
func (*PullServiceConfigReq) ProtoMessage()    {}
func (*PullServiceConfigReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_base_protocol_013846fdbab5dc65, []int{7}
}
func (m *PullServiceConfigReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PullServiceConfigReq.Unmarshal(m, b)
}
func (m *PullServiceConfigReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PullServiceConfigReq.Marshal(b, m, deterministic)
}
func (dst *PullServiceConfigReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PullServiceConfigReq.Merge(dst, src)
}
func (m *PullServiceConfigReq) XXX_Size() int {
	return xxx_messageInfo_PullServiceConfigReq.Size(m)
}
func (m *PullServiceConfigReq) XXX_DiscardUnknown() {
	xxx_messageInfo_PullServiceConfigReq.DiscardUnknown(m)
}

var xxx_messageInfo_PullServiceConfigReq proto.InternalMessageInfo

func (m *PullServiceConfigReq) GetServiceKey() string {
	if m != nil && m.ServiceKey != nil {
		return *m.ServiceKey
	}
	return ""
}

func (m *PullServiceConfigReq) GetVersion() uint32 {
	if m != nil && m.Version != nil {
		return *m.Version
	}
	return 0
}

// broker回复agent一个服务的最新配置
type PullServiceConfigRsp struct {
	ServiceKey           *string  `protobuf:"bytes,1,req,name=serviceKey" json:"serviceKey,omitempty"`
	Version              *uint32  `protobuf:"varint,2,req,name=version" json:"version,omitempty"`
	ConfKeys             []string `protobuf:"bytes,3,rep,name=confKeys" json:"confKeys,omitempty"`
	Values               []string `protobuf:"bytes,4,rep,name=values" json:"values,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PullServiceConfigRsp) Reset()         { *m = PullServiceConfigRsp{} }
func (m *PullServiceConfigRsp) String() string { return proto.CompactTextString(m) }
func (*PullServiceConfigRsp) ProtoMessage()    {}
func (*PullServiceConfigRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_base_protocol_013846fdbab5dc65, []int{8}
}
func (m *PullServiceConfigRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PullServiceConfigRsp.Unmarshal(m, b)
}
func (m *PullServiceConfigRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PullServiceConfigRsp.Marshal(b, m, deterministic)
}
func (dst *PullServiceConfigRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PullServiceConfigRsp.Merge(dst, src)
}
func (m *PullServiceConfigRsp) XXX_Size() int {
	return xxx_messageInfo_PullServiceConfigRsp.Size(m)
}
func (m *PullServiceConfigRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_PullServiceConfigRsp.DiscardUnknown(m)
}

var xxx_messageInfo_PullServiceConfigRsp proto.InternalMessageInfo

func (m *PullServiceConfigRsp) GetServiceKey() string {
	if m != nil && m.ServiceKey != nil {
		return *m.ServiceKey
	}
	return ""
}

func (m *PullServiceConfigRsp) GetVersion() uint32 {
	if m != nil && m.Version != nil {
		return *m.Version
	}
	return 0
}

func (m *PullServiceConfigRsp) GetConfKeys() []string {
	if m != nil {
		return m.ConfKeys
	}
	return nil
}

func (m *PullServiceConfigRsp) GetValues() []string {
	if m != nil {
		return m.Values
	}
	return nil
}

func init() {
	proto.RegisterType((*KeepUsingReq)(nil), "protocol.KeepUsingReq")
	proto.RegisterType((*SubscribeReq)(nil), "protocol.SubscribeReq")
	proto.RegisterType((*SubscribeBrokerRsp)(nil), "protocol.SubscribeBrokerRsp")
	proto.RegisterType((*SubscribeAgentRsp)(nil), "protocol.SubscribeAgentRsp")
	proto.RegisterType((*AddConfigReq)(nil), "protocol.AddConfigReq")
	proto.RegisterType((*DelConfigReq)(nil), "protocol.DelConfigReq")
	proto.RegisterType((*UpdateConfigReq)(nil), "protocol.UpdateConfigReq")
	proto.RegisterType((*PullServiceConfigReq)(nil), "protocol.PullServiceConfigReq")
	proto.RegisterType((*PullServiceConfigRsp)(nil), "protocol.PullServiceConfigRsp")
	proto.RegisterEnum("protocol.MsgTypeId", MsgTypeId_name, MsgTypeId_value)
}

func init() { proto.RegisterFile("base_protocol.proto", fileDescriptor_base_protocol_013846fdbab5dc65) }

var fileDescriptor_base_protocol_013846fdbab5dc65 = []byte{
	// 355 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xac, 0x52, 0xcf, 0x6b, 0xea, 0x40,
	0x10, 0xc6, 0xfc, 0x50, 0x33, 0x44, 0x5d, 0x47, 0xdf, 0x7b, 0xcb, 0x3b, 0x85, 0x9c, 0xc2, 0x3b,
	0xbc, 0x5b, 0x4f, 0x3d, 0xd9, 0x7a, 0x59, 0xa4, 0x50, 0xb4, 0xf6, 0x56, 0x8a, 0x66, 0x47, 0x09,
	0x0d, 0xd9, 0x34, 0xab, 0x82, 0x7f, 0x6c, 0xff, 0x97, 0x92, 0x6d, 0x2b, 0x41, 0x23, 0xb4, 0xb4,
	0xb7, 0xcc, 0xe4, 0xdb, 0x6f, 0xbe, 0x99, 0xef, 0x83, 0xc1, 0x72, 0xa1, 0xe9, 0x31, 0x2f, 0xd4,
	0x46, 0xc5, 0x2a, 0xfd, 0x6f, 0x3e, 0xb0, 0xfd, 0x51, 0x87, 0x21, 0xf8, 0x13, 0xa2, 0x7c, 0xae,
	0x93, 0x6c, 0x3d, 0xa5, 0x67, 0x44, 0x00, 0x4d, 0xc5, 0x2e, 0x89, 0x69, 0x42, 0x7b, 0xde, 0x08,
	0xac, 0xc8, 0x2b, 0x31, 0xb3, 0xed, 0x52, 0xc7, 0x45, 0xb2, 0xa4, 0x73, 0x98, 0x04, 0xf0, 0x80,
	0xb9, 0x2a, 0xd4, 0x13, 0x15, 0x53, 0x9d, 0xd7, 0x21, 0xd1, 0x07, 0x27, 0x56, 0x92, 0xb8, 0x15,
	0x58, 0x91, 0x8b, 0x3d, 0x68, 0xed, 0xa8, 0xd0, 0x89, 0xca, 0xb8, 0x1d, 0x58, 0x51, 0x07, 0x19,
	0xb4, 0x63, 0x95, 0xad, 0x26, 0xb4, 0xd7, 0xdc, 0x09, 0xec, 0xc8, 0xc3, 0x2e, 0x34, 0x77, 0x8b,
	0x74, 0x4b, 0x9a, 0xbb, 0x65, 0x1d, 0x5e, 0x40, 0xff, 0x30, 0x6a, 0xb4, 0xa6, 0x6c, 0xf3, 0xa9,
	0x49, 0xe1, 0x0c, 0xfc, 0x91, 0x94, 0xd7, 0x2a, 0x5b, 0x25, 0xe7, 0x36, 0xad, 0xaa, 0xb1, 0x8c,
	0x9a, 0x1e, 0xb4, 0xde, 0xd5, 0x18, 0x79, 0x1e, 0x76, 0xc0, 0x35, 0x62, 0xb8, 0x63, 0xd6, 0x1e,
	0x83, 0x3f, 0xa6, 0xf4, 0x9b, 0xa4, 0xe1, 0x3d, 0xf4, 0xe6, 0xb9, 0x5c, 0x6c, 0xe8, 0x87, 0xd5,
	0x5d, 0xc2, 0xf0, 0x76, 0x9b, 0xa6, 0xb3, 0x37, 0xa2, 0xaf, 0x91, 0x87, 0x0f, 0x75, 0x8f, 0xcf,
	0x5c, 0xfa, 0x44, 0x59, 0xd5, 0x45, 0xfb, 0xc8, 0x45, 0xe3, 0xea, 0xbf, 0x97, 0x06, 0x78, 0x37,
	0x7a, 0x7d, 0xb7, 0xcf, 0x49, 0x48, 0x44, 0xe8, 0x56, 0x63, 0x28, 0x24, 0x6b, 0x94, 0xbd, 0x6a,
	0xec, 0x84, 0x64, 0x16, 0x72, 0x18, 0x9e, 0xc6, 0x4c, 0x48, 0x66, 0xe3, 0x1f, 0x18, 0x9c, 0xa4,
	0x42, 0x48, 0xe6, 0x94, 0x34, 0x55, 0xdf, 0x85, 0x64, 0x6e, 0xd9, 0xab, 0xda, 0x26, 0x24, 0x6b,
	0xe2, 0x2f, 0xe8, 0x1f, 0x99, 0x20, 0x24, 0x6b, 0xe1, 0x5f, 0xf8, 0x5d, 0x77, 0x43, 0x21, 0x59,
	0xbb, 0xfe, 0x9f, 0x19, 0xeb, 0xbd, 0x06, 0x00, 0x00, 0xff, 0xff, 0x98, 0xd0, 0x8d, 0x62, 0x78,
	0x03, 0x00, 0x00,
}
