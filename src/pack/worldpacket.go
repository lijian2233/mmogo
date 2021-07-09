package pack

import (
	"errors"
	"fmt"
	"mmogo/common"
	"mmogo/protocal/binary"
	"unsafe"
)

const packetHeaderSize uint32 = 10
const maxContentSize = uint32(0x01000000)

var MaxSizeError = errors.New("excced max error")
var NumTypeError = errors.New("num type error")
var WriteAcessError = errors.New("write access error")

/*
*content matain 9 bytes
* 4bytes seqno
* 2bytes op
* 3bytes packet length
 */
type worldPacket struct {
	content    []byte //message body
	seqNo      *uint32
	op         *uint16
	readIndex  uint32
	writeIndex uint32
	contentCap uint32
}


func GetWorldPacketHeaderSize() uint32 {
	return packetHeaderSize
}

func NewWorldPacket(op uint16, cap uint32) *worldPacket {
	packet := new(worldPacket)
	if cap >= maxContentSize {
		panic(fmt.Sprintf("packet must < %x", maxContentSize))
	}

	packet.readIndex = packetHeaderSize
	packet.writeIndex = packetHeaderSize
	packet.content = make([]byte, cap+packetHeaderSize, cap+packetHeaderSize)
	packet.seqNo = (*uint32)(unsafe.Pointer(&packet.content[0]))
	packet.op = (*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(&packet.content[0])) + uintptr(4)))
	packet.contentCap = cap + packetHeaderSize
	*packet.op = op
	packet.setSeqNo(0)
	return packet
}

func ParsePacketHeader(header []byte) (seqNo, bodySize uint32, op uint16, ok bool) {
	if header == nil || uint32(len(header)) < packetHeaderSize {
		return 0,0, 0, false
	}

	seqNo = *(*uint32)(unsafe.Pointer(&header[0]))
	op = *(*uint16)(unsafe.Pointer(&header[4]))
	bodySize = *(*uint32)(unsafe.Pointer(&header[6]))

	if common.IsLittleEnd(){
		seqNo = binary.SwapUint32(seqNo)
		op = binary.SwapUint16(op)
		bodySize = binary.SwapUint32(bodySize)
	}
	return seqNo, bodySize, op,true
}

func (packet* worldPacket) getReadIndex() uint32 {
	return packet.readIndex
}

func (packet* worldPacket) getWriteIndex() uint32  {
	return packet.writeIndex
}

func (packet *worldPacket) Reset(op uint16) {
	packet.readIndex = packetHeaderSize
	packet.writeIndex = packetHeaderSize
	if common.IsLittleEnd() {
		*packet.op = binary.SwapUint16(op)
	}else{
		*packet.op = op
	}
}

func (packet *worldPacket) setSeqNo(sNo uint32) {
	if common.IsLittleEnd() {
		*packet.seqNo = binary.SwapUint32(sNo)
	}else{
		*packet.seqNo = sNo
	}
}

func (packet *worldPacket) checkPacketSize(len uint32) bool {
	if packet.writeIndex+len > maxContentSize {
		return false
	}

	if packet.writeIndex+len >= packet.contentCap {
		nLen := 2*(packet.contentCap+len)
		if nLen > maxContentSize+packetHeaderSize {
			nLen = maxContentSize + packetHeaderSize
		}
		newContent := make([]byte, nLen, nLen)
		fmt.Println(packet.content)
		copy(newContent[0:], packet.content[0:])
		packet.contentCap = nLen
		packet.content = newContent
		packet.seqNo = (*uint32)(unsafe.Pointer(&newContent[0]))
		packet.op = (*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(&newContent[0])) + uintptr(4)))
	}
	return true
}

func (packet *worldPacket) getOpcode() uint16 {
	return *packet.op
}

func (packet *worldPacket) getSeqNo() uint32 {
	return *packet.seqNo
}

func (packet *worldPacket) WriteBytes(b []byte) (uint32, error) {
	nLen := uint32(len(b))
	if !packet.checkPacketSize(nLen) {
		return packet.writeIndex, MaxSizeError
	}
	copy(packet.content[packet.writeIndex:], b[0:])

	packet.writeIndex += nLen
	return packet.writeIndex, nil
}

func (packet *worldPacket) WriteString(str string) (uint32, error) {
	len := len(str)
	if len >= 0x01000000 {
		return packet.writeIndex, MaxSizeError
	}

	if !packet.checkPacketSize(uint32(len) + 3) {
		return packet.writeIndex, MaxSizeError
	}

	bb := [4]byte{}
	binary.Uint32ToBytes(uint32(len), &bb)
	for i := 1; i < 4; i++ {
		packet.content[packet.writeIndex+uint32(i-1)] = bb[i]
	}
	packet.writeIndex += 3

	b := binary.String2Byte(str)
	return packet.WriteBytes(b)
}

func (packet *worldPacket) readString() string {
	if packet.readIndex+3 < packet.contentCap+packetHeaderSize {
		b := [4]byte{0x00,}
		copy(b[1:], packet.content[packet.readIndex:])
		packet.readIndex += 3
		l := binary.BytesToUint32(b)
		str := string(packet.content[packet.readIndex : packet.readIndex+l])
		packet.readIndex += l
		return str
	}
	return ""
}

func (packet *worldPacket) WriteByte(b byte, wIndex ...uint32) (uint32, error) {
	writeIndex := packet.writeIndex
	if wIndex != nil {
		writeIndex = wIndex[0]
		if writeIndex >= packet.contentCap{
			return packet.writeIndex, MaxSizeError
		}
		packet.content[writeIndex] = b
	}else {
		if !packet.checkPacketSize(1) {
			return packet.writeIndex, MaxSizeError
		}
		packet.content[writeIndex] = b
		packet.writeIndex += 1
	}
	return packet.writeIndex, nil
}

func getIntTypeSize(num interface{}) uint32 {
	switch num.(type) {
	case int32,uint32:return 4
	case int16,uint16:return 2
	case int8,uint8:return 1
	case int,uint,int64,uint64:return 8
	default:
		return 0
	}
}

func (packet *worldPacket) WriteNum(num interface{}, wIndex ...uint32) (uint32, error) {
	writeIndex := packet.writeIndex
	if wIndex != nil{
		writeIndex = wIndex[0]
		if writeIndex >= packet.contentCap{
			return packet.writeIndex, MaxSizeError
		}
	}

	l:=getIntTypeSize(num)
	if l == 0 {
		return packet.writeIndex, NumTypeError
	}

	if wIndex != nil{
		if writeIndex + l >= packet.contentCap{
			return packet.writeIndex, WriteAcessError
		}
	}else if !packet.checkPacketSize(l) {
		return packet.writeIndex, MaxSizeError
	}

	switch num.(type) {
	case int8:
		val, _ := num.(int8)
		return packet.WriteByte(byte(val), wIndex...)
	case uint8:
		val, _ := num.(uint8)
		return packet.WriteByte(byte(val), wIndex...)
	case int16:
		val, _ := num.(int16)
		b := [2]byte{}
		binary.Int16ToByte(val, &b)
		for i := 0; i < 2; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}

		if wIndex == nil {
			packet.writeIndex += 2
		}
		return packet.writeIndex, nil

	case uint16:
		val, _ := num.(uint16)
		b := [2]byte{}
		binary.Uint16ToByte(val, &b)
		for i := 0; i < 2; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 2
		}
		return packet.writeIndex, nil
	case int32:
		val, _ := num.(int32)
		b := [4]byte{}
		binary.Int32ToBytes(val, &b)
		for i := 0; i < 4; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 4
		}
		return packet.writeIndex, nil
	case uint32:
		val, _ := num.(uint32)
		b := [4]byte{}
		binary.Uint32ToBytes(val, &b)
		for i := 0; i < 4; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 4
		}
		return packet.writeIndex, nil
	case int64:
		b := [8]byte{}
		val, _ := num.(int64)
		binary.Int64ToBytes(val, &b)
		for i := 0; i < 8; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 8
		}
		return packet.writeIndex, nil
	case uint64:
		val, _ := num.(uint64)
		b := [8]byte{}
		binary.Uint64ToBytes(val, &b)
		for i := 0; i < 8; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 8
		}
		return packet.writeIndex, nil
	case int:
		b := [8]byte{}
		val, _ := num.(int)
		binary.Int64ToBytes(int64(val), &b)
		for i := 0; i < 8; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 8
		}
		return packet.writeIndex, nil
	case uint:
		b := [8]byte{}
		val, _ := num.(uint)
		binary.Int64ToBytes(int64(val), &b)
		for i := 0; i < 8; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 8
		}
		return packet.writeIndex, nil
	default:
		return packet.writeIndex, NumTypeError
	}
}

func (packet *worldPacket) ReadInt8() int8 {
	if packet.readIndex+1 < packet.contentCap {
		packet.readIndex += 1
		return int8(packet.content[packet.readIndex])
	}
	packet.readIndex = packet.contentCap
	return 0
}

func (packet *worldPacket) ReadUint8() uint8 {
	return uint8(packet.ReadInt8())
}

func (packet *worldPacket) ReadInt16() int16 {
	if packet.readIndex+2 < packet.contentCap {
		ret := *(*int16)(unsafe.Pointer(&packet.content[packet.readIndex]))
		packet.readIndex += 2
		if common.IsLittleEnd(){
			return binary.SwapInt16(ret)
		}
		return ret
	}
	packet.readIndex = packet.contentCap
	return 0
}

func (packet *worldPacket) ReadUint16() uint16 {
	return uint16(packet.ReadInt16())
}

func (packet *worldPacket) ReadInt32() int32 {
	if packet.readIndex+4 < packet.contentCap {
		ret := *(*int32)(unsafe.Pointer(&packet.content[packet.readIndex]))
		packet.readIndex += 4
		if common.IsLittleEnd(){
			return binary.SwapInt32(ret)
		}
		return ret
	}
	packet.readIndex = packet.contentCap
	return 0
}

func (packet *worldPacket) ReadUint32() uint32 {
	return uint32(packet.ReadInt32())
}

func (packet *worldPacket) ReadInt64() int64 {
	if packet.readIndex+8 < packet.contentCap {
		ret := *(*int64)(unsafe.Pointer(&packet.content[packet.readIndex]))
		packet.readIndex += 8

		if common.IsLittleEnd(){
			return binary.SwapInt64(ret)
		}
		return ret
	}
	packet.readIndex = packet.contentCap
	return 0
}

func (packet *worldPacket) ReadUInt64() uint64 {
	return uint64(packet.ReadInt64())
}

func (packet *worldPacket) ReadInt() int {
	return int(packet.ReadInt64())
}

func (packet *worldPacket) ReadUint() uint {
	return uint(packet.ReadInt64())
}

func (packet *worldPacket) WriteBool(b bool) (uint32, error) {
	bb := byte(0)
	if b {
		bb = 1
	} else {
		bb = 0
	}
	return packet.WriteByte(bb)
}

func (packet *worldPacket) ReadBool() bool {
	val := packet.ReadUint8()
	return val != 0
}
