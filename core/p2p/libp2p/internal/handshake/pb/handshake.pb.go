// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: handshake.proto

package pb

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type Syn struct {
	ObservedUnderlay []byte `protobuf:"bytes,1,opt,name=ObservedUnderlay,proto3" json:"ObservedUnderlay,omitempty"`
}

func (m *Syn) Reset()         { *m = Syn{} }
func (m *Syn) String() string { return proto.CompactTextString(m) }
func (*Syn) ProtoMessage()    {}
func (*Syn) Descriptor() ([]byte, []int) {
	return fileDescriptor_a77305914d5d202f, []int{0}
}
func (m *Syn) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Syn) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Syn.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Syn) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Syn.Merge(m, src)
}
func (m *Syn) XXX_Size() int {
	return m.Size()
}
func (m *Syn) XXX_DiscardUnknown() {
	xxx_messageInfo_Syn.DiscardUnknown(m)
}

var xxx_messageInfo_Syn proto.InternalMessageInfo

func (m *Syn) GetObservedUnderlay() []byte {
	if m != nil {
		return m.ObservedUnderlay
	}
	return nil
}

type Ack struct {
	Address        *MopAddress `protobuf:"bytes,1,opt,name=Address,proto3" json:"Address,omitempty"`
	NetworkID      uint64      `protobuf:"varint,2,opt,name=NetworkID,proto3" json:"NetworkID,omitempty"`
	FullNode       bool        `protobuf:"varint,3,opt,name=FullNode,proto3" json:"FullNode,omitempty"`
	Nonce          []byte      `protobuf:"bytes,4,opt,name=Nonce,proto3" json:"Nonce,omitempty"`
	WelcomeMessage string      `protobuf:"bytes,99,opt,name=WelcomeMessage,proto3" json:"WelcomeMessage,omitempty"`
}

func (m *Ack) Reset()         { *m = Ack{} }
func (m *Ack) String() string { return proto.CompactTextString(m) }
func (*Ack) ProtoMessage()    {}
func (*Ack) Descriptor() ([]byte, []int) {
	return fileDescriptor_a77305914d5d202f, []int{1}
}
func (m *Ack) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Ack) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Ack.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Ack) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Ack.Merge(m, src)
}
func (m *Ack) XXX_Size() int {
	return m.Size()
}
func (m *Ack) XXX_DiscardUnknown() {
	xxx_messageInfo_Ack.DiscardUnknown(m)
}

var xxx_messageInfo_Ack proto.InternalMessageInfo

func (m *Ack) GetAddress() *MopAddress {
	if m != nil {
		return m.Address
	}
	return nil
}

func (m *Ack) GetNetworkID() uint64 {
	if m != nil {
		return m.NetworkID
	}
	return 0
}

func (m *Ack) GetFullNode() bool {
	if m != nil {
		return m.FullNode
	}
	return false
}

func (m *Ack) GetNonce() []byte {
	if m != nil {
		return m.Nonce
	}
	return nil
}

func (m *Ack) GetWelcomeMessage() string {
	if m != nil {
		return m.WelcomeMessage
	}
	return ""
}

type SynAck struct {
	Syn *Syn `protobuf:"bytes,1,opt,name=Syn,proto3" json:"Syn,omitempty"`
	Ack *Ack `protobuf:"bytes,2,opt,name=Ack,proto3" json:"Ack,omitempty"`
}

func (m *SynAck) Reset()         { *m = SynAck{} }
func (m *SynAck) String() string { return proto.CompactTextString(m) }
func (*SynAck) ProtoMessage()    {}
func (*SynAck) Descriptor() ([]byte, []int) {
	return fileDescriptor_a77305914d5d202f, []int{2}
}
func (m *SynAck) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SynAck) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SynAck.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SynAck) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SynAck.Merge(m, src)
}
func (m *SynAck) XXX_Size() int {
	return m.Size()
}
func (m *SynAck) XXX_DiscardUnknown() {
	xxx_messageInfo_SynAck.DiscardUnknown(m)
}

var xxx_messageInfo_SynAck proto.InternalMessageInfo

func (m *SynAck) GetSyn() *Syn {
	if m != nil {
		return m.Syn
	}
	return nil
}

func (m *SynAck) GetAck() *Ack {
	if m != nil {
		return m.Ack
	}
	return nil
}

type MopAddress struct {
	Underlay  []byte `protobuf:"bytes,1,opt,name=Underlay,proto3" json:"Underlay,omitempty"`
	Signature []byte `protobuf:"bytes,2,opt,name=Signature,proto3" json:"Signature,omitempty"`
	Overlay   []byte `protobuf:"bytes,3,opt,name=Overlay,proto3" json:"Overlay,omitempty"`
}

func (m *MopAddress) Reset()         { *m = MopAddress{} }
func (m *MopAddress) String() string { return proto.CompactTextString(m) }
func (*MopAddress) ProtoMessage()    {}
func (*MopAddress) Descriptor() ([]byte, []int) {
	return fileDescriptor_a77305914d5d202f, []int{3}
}
func (m *MopAddress) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MopAddress) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MopAddress.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MopAddress) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MopAddress.Merge(m, src)
}
func (m *MopAddress) XXX_Size() int {
	return m.Size()
}
func (m *MopAddress) XXX_DiscardUnknown() {
	xxx_messageInfo_MopAddress.DiscardUnknown(m)
}

var xxx_messageInfo_MopAddress proto.InternalMessageInfo

func (m *MopAddress) GetUnderlay() []byte {
	if m != nil {
		return m.Underlay
	}
	return nil
}

func (m *MopAddress) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *MopAddress) GetOverlay() []byte {
	if m != nil {
		return m.Overlay
	}
	return nil
}

func init() {
	proto.RegisterType((*Syn)(nil), "handshake.Syn")
	proto.RegisterType((*Ack)(nil), "handshake.Ack")
	proto.RegisterType((*SynAck)(nil), "handshake.SynAck")
	proto.RegisterType((*MopAddress)(nil), "handshake.MopAddress")
}

func init() { proto.RegisterFile("handshake.proto", fileDescriptor_a77305914d5d202f) }

var fileDescriptor_a77305914d5d202f = []byte{
	// 317 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x91, 0xdf, 0x4a, 0xc3, 0x30,
	0x14, 0xc6, 0x97, 0x75, 0xee, 0xcf, 0x71, 0x4c, 0x09, 0x0a, 0x45, 0x46, 0x29, 0xbd, 0x90, 0xe2,
	0xc5, 0x44, 0x7d, 0x82, 0x89, 0x08, 0x82, 0xdb, 0x20, 0x45, 0x04, 0xaf, 0xec, 0x9a, 0xc3, 0x26,
	0xad, 0xc9, 0x48, 0xb6, 0x49, 0xdf, 0xc2, 0x27, 0xf1, 0x39, 0xbc, 0xdc, 0xa5, 0x97, 0xb2, 0xbd,
	0x88, 0x34, 0xdb, 0x5a, 0x9d, 0x97, 0xdf, 0x77, 0xbe, 0x24, 0xbf, 0xef, 0x04, 0x0e, 0xc6, 0xa1,
	0xe0, 0x7a, 0x1c, 0xc6, 0xd8, 0x99, 0x28, 0x39, 0x95, 0xb4, 0x91, 0x1b, 0xde, 0x05, 0x58, 0x41,
	0x2a, 0xe8, 0x19, 0x1c, 0x0e, 0x86, 0x1a, 0xd5, 0x1c, 0xf9, 0x83, 0xe0, 0xa8, 0x92, 0x30, 0xb5,
	0x89, 0x4b, 0xfc, 0x26, 0xfb, 0xe7, 0x7b, 0x1f, 0x04, 0xac, 0x6e, 0x14, 0xd3, 0x73, 0xa8, 0x75,
	0x39, 0x57, 0xa8, 0xb5, 0x89, 0xee, 0x5f, 0x1e, 0x77, 0x8a, 0x87, 0x7a, 0x72, 0xb2, 0x19, 0xb2,
	0x6d, 0x8a, 0xb6, 0xa1, 0xd1, 0xc7, 0xe9, 0x9b, 0x54, 0xf1, 0xdd, 0x8d, 0x5d, 0x76, 0x89, 0x5f,
	0x61, 0x85, 0x41, 0x4f, 0xa0, 0x7e, 0x3b, 0x4b, 0x92, 0xbe, 0xe4, 0x68, 0x5b, 0x2e, 0xf1, 0xeb,
	0x2c, 0xd7, 0xf4, 0x08, 0xf6, 0xfa, 0x52, 0x44, 0x68, 0x57, 0x0c, 0xd3, 0x5a, 0xd0, 0x53, 0x68,
	0x3d, 0x62, 0x12, 0xc9, 0x57, 0xec, 0xa1, 0xd6, 0xe1, 0x08, 0xed, 0xc8, 0x25, 0x7e, 0x83, 0xed,
	0xb8, 0xde, 0x3d, 0x54, 0x83, 0x54, 0x64, 0xc8, 0xae, 0x69, 0xbb, 0xc1, 0x6d, 0xfd, 0xc2, 0x0d,
	0x52, 0xc1, 0xcc, 0x22, 0x5c, 0xd3, 0xcd, 0xd0, 0xfd, 0x4d, 0x74, 0xa3, 0x98, 0x65, 0x23, 0xef,
	0x19, 0xa0, 0x28, 0x97, 0x51, 0xef, 0x2c, 0x2c, 0xd7, 0x59, 0xdf, 0xe0, 0x65, 0x24, 0xc2, 0xe9,
	0x4c, 0xa1, 0xb9, 0xb1, 0xc9, 0x0a, 0x83, 0xda, 0x50, 0x1b, 0xcc, 0xd7, 0x07, 0x2d, 0x33, 0xdb,
	0xca, 0xeb, 0xf6, 0xe7, 0xd2, 0x21, 0x8b, 0xa5, 0x43, 0xbe, 0x97, 0x0e, 0x79, 0x5f, 0x39, 0xa5,
	0xc5, 0xca, 0x29, 0x7d, 0xad, 0x9c, 0xd2, 0x53, 0x79, 0x32, 0x1c, 0x56, 0xcd, 0x1f, 0x5e, 0xfd,
	0x04, 0x00, 0x00, 0xff, 0xff, 0xb4, 0x91, 0xcc, 0x96, 0xd6, 0x01, 0x00, 0x00,
}

func (m *Syn) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Syn) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Syn) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ObservedUnderlay) > 0 {
		i -= len(m.ObservedUnderlay)
		copy(dAtA[i:], m.ObservedUnderlay)
		i = encodeVarintHandshake(dAtA, i, uint64(len(m.ObservedUnderlay)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Ack) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Ack) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Ack) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.WelcomeMessage) > 0 {
		i -= len(m.WelcomeMessage)
		copy(dAtA[i:], m.WelcomeMessage)
		i = encodeVarintHandshake(dAtA, i, uint64(len(m.WelcomeMessage)))
		i--
		dAtA[i] = 0x6
		i--
		dAtA[i] = 0x9a
	}
	if len(m.Nonce) > 0 {
		i -= len(m.Nonce)
		copy(dAtA[i:], m.Nonce)
		i = encodeVarintHandshake(dAtA, i, uint64(len(m.Nonce)))
		i--
		dAtA[i] = 0x22
	}
	if m.FullNode {
		i--
		if m.FullNode {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x18
	}
	if m.NetworkID != 0 {
		i = encodeVarintHandshake(dAtA, i, uint64(m.NetworkID))
		i--
		dAtA[i] = 0x10
	}
	if m.Address != nil {
		{
			size, err := m.Address.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintHandshake(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *SynAck) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SynAck) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SynAck) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Ack != nil {
		{
			size, err := m.Ack.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintHandshake(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if m.Syn != nil {
		{
			size, err := m.Syn.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintHandshake(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MopAddress) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MopAddress) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MopAddress) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Overlay) > 0 {
		i -= len(m.Overlay)
		copy(dAtA[i:], m.Overlay)
		i = encodeVarintHandshake(dAtA, i, uint64(len(m.Overlay)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Signature) > 0 {
		i -= len(m.Signature)
		copy(dAtA[i:], m.Signature)
		i = encodeVarintHandshake(dAtA, i, uint64(len(m.Signature)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Underlay) > 0 {
		i -= len(m.Underlay)
		copy(dAtA[i:], m.Underlay)
		i = encodeVarintHandshake(dAtA, i, uint64(len(m.Underlay)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintHandshake(dAtA []byte, offset int, v uint64) int {
	offset -= sovHandshake(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Syn) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ObservedUnderlay)
	if l > 0 {
		n += 1 + l + sovHandshake(uint64(l))
	}
	return n
}

func (m *Ack) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Address != nil {
		l = m.Address.Size()
		n += 1 + l + sovHandshake(uint64(l))
	}
	if m.NetworkID != 0 {
		n += 1 + sovHandshake(uint64(m.NetworkID))
	}
	if m.FullNode {
		n += 2
	}
	l = len(m.Nonce)
	if l > 0 {
		n += 1 + l + sovHandshake(uint64(l))
	}
	l = len(m.WelcomeMessage)
	if l > 0 {
		n += 2 + l + sovHandshake(uint64(l))
	}
	return n
}

func (m *SynAck) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Syn != nil {
		l = m.Syn.Size()
		n += 1 + l + sovHandshake(uint64(l))
	}
	if m.Ack != nil {
		l = m.Ack.Size()
		n += 1 + l + sovHandshake(uint64(l))
	}
	return n
}

func (m *MopAddress) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Underlay)
	if l > 0 {
		n += 1 + l + sovHandshake(uint64(l))
	}
	l = len(m.Signature)
	if l > 0 {
		n += 1 + l + sovHandshake(uint64(l))
	}
	l = len(m.Overlay)
	if l > 0 {
		n += 1 + l + sovHandshake(uint64(l))
	}
	return n
}

func sovHandshake(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozHandshake(x uint64) (n int) {
	return sovHandshake(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Syn) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowHandshake
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Syn: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Syn: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObservedUnderlay", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHandshake
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthHandshake
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthHandshake
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ObservedUnderlay = append(m.ObservedUnderlay[:0], dAtA[iNdEx:postIndex]...)
			if m.ObservedUnderlay == nil {
				m.ObservedUnderlay = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipHandshake(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthHandshake
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Ack) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowHandshake
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Ack: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Ack: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHandshake
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthHandshake
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthHandshake
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Address == nil {
				m.Address = &MopAddress{}
			}
			if err := m.Address.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NetworkID", wireType)
			}
			m.NetworkID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHandshake
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NetworkID |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field FullNode", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHandshake
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.FullNode = bool(v != 0)
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Nonce", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHandshake
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthHandshake
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthHandshake
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Nonce = append(m.Nonce[:0], dAtA[iNdEx:postIndex]...)
			if m.Nonce == nil {
				m.Nonce = []byte{}
			}
			iNdEx = postIndex
		case 99:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field WelcomeMessage", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHandshake
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHandshake
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthHandshake
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.WelcomeMessage = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipHandshake(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthHandshake
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SynAck) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowHandshake
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SynAck: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SynAck: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Syn", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHandshake
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthHandshake
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthHandshake
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Syn == nil {
				m.Syn = &Syn{}
			}
			if err := m.Syn.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ack", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHandshake
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthHandshake
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthHandshake
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Ack == nil {
				m.Ack = &Ack{}
			}
			if err := m.Ack.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipHandshake(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthHandshake
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MopAddress) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowHandshake
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MopAddress: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MopAddress: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Underlay", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHandshake
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthHandshake
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthHandshake
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Underlay = append(m.Underlay[:0], dAtA[iNdEx:postIndex]...)
			if m.Underlay == nil {
				m.Underlay = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signature", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHandshake
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthHandshake
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthHandshake
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signature = append(m.Signature[:0], dAtA[iNdEx:postIndex]...)
			if m.Signature == nil {
				m.Signature = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Overlay", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHandshake
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthHandshake
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthHandshake
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Overlay = append(m.Overlay[:0], dAtA[iNdEx:postIndex]...)
			if m.Overlay == nil {
				m.Overlay = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipHandshake(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthHandshake
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipHandshake(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowHandshake
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowHandshake
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowHandshake
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthHandshake
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupHandshake
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthHandshake
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthHandshake        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowHandshake          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupHandshake = fmt.Errorf("proto: unexpected end of group")
)
