package packet

import (
	"errors"
	"fmt"
	"mmogo/interface"
	"mmogo/lib/binaryop"
	"mmogo/lib/util"
	"unsafe"
)

const packetHeaderSize uint32 = 10
const maxContentSize = uint32(0x01000000)
const seqOffset = 0
const opOffset = 4
const sizeOffset = 6

var MaxSizeError = errors.New("excced max error")
var NumTypeError = errors.New("num type error")
var WriteAcessError = errors.New("write access error")
var BuildPacketError = errors.New("build packet error")

/*
*content matain 9 bytes
* 4bytes seqno
* 2bytes op
* 3bytes packet length
 */
type UtilPacket struct {
	content    []byte //message body
	seqNo      *uint32
	op         *uint16
	size       *uint32
	readIndex  uint32
	writeIndex uint32
	contentCap uint32
}

func (packet* UtilPacket) BodySize() uint32 {
	return packet.writeIndex - packet.GetSize()
}

func (packet* UtilPacket) HeaderSize() uint32  {
	return packetHeaderSize
}

func (packet* UtilPacket) GetBody(safe ...interface{}) []byte {
	bSafe := true
	if len(safe) > 0 {
		v, ok := safe[0].(bool)
		if ok {
			bSafe = v
		}
	}
	if packet.BodySize() == 0 {
		return nil
	}

	if bSafe {
		body := make([]byte, packet.BodySize(), packet.BodySize())
		copy(body, packet.content[packet.HeaderSize():])
		return body
	}else{
		h := [3]uintptr{uintptr(unsafe.Pointer(&packet.content[packetHeaderSize])), uintptr(packet.BodySize()), uintptr(packet.BodySize())}
		return *(*[]byte)(unsafe.Pointer(&h))
	}
}

func (packet* UtilPacket) GetHeader(safe ...interface{}) []byte  {
	bSafe := true
	if len(safe) > 0 {
		v, ok := safe[0].(bool)
		if ok {
			bSafe = v
		}
	}

	if bSafe {
		header := make([]byte, packet.HeaderSize(), packet.HeaderSize())
		copy(header, packet.content[0:packet.HeaderSize()])
		return header
	}else{
		h := [3]uintptr{uintptr(unsafe.Pointer(&packet.content[0])), uintptr(packet.HeaderSize()), uintptr(packet.HeaderSize())}
		return *(*[]byte)(unsafe.Pointer(&h))
	}
}

func (packet* UtilPacket) GetPacket(safe ...interface{}) []byte  {
	bSafe := true
	if len(safe) > 0 {
		v, ok := safe[0].(bool)
		if ok {
			bSafe = v
		}
	}

	if bSafe {
		full := make([]byte, packet.writeIndex, packet.writeIndex)
		copy(full, packet.content[0:packet.writeIndex])
		return full
	}else{
		h := [3]uintptr{uintptr(unsafe.Pointer(&packet.content[0])), uintptr(packet.writeIndex), uintptr(packet.writeIndex)}
		return *(*[]byte)(unsafe.Pointer(&h))
	}
}

func NewWorldPacket(op uint16, cap uint32) *UtilPacket {
	packet := new(UtilPacket)
	if cap >= maxContentSize {
		panic(fmt.Sprintf("packet must < %x", maxContentSize))
	}

	packet.readIndex = packetHeaderSize
	packet.writeIndex = packetHeaderSize
	packet.content = make([]byte, cap+packetHeaderSize, cap+packetHeaderSize)
	packet.seqNo = (*uint32)(unsafe.Pointer(&packet.content[seqOffset]))
	packet.op = (*uint16)(unsafe.Pointer(unsafe.Pointer(&packet.content[opOffset])))
	packet.size = (*uint32)(unsafe.Pointer(unsafe.Pointer(&packet.content[sizeOffset])))
	packet.contentCap = cap + packetHeaderSize
	packet.SetSeqNo(0)
	packet.SetSize(packet.writeIndex)
	packet.SetOp(op)
	return packet
}

func (packet* UtilPacket) PacketSize() uint32 {
	return packet.GetSize()
}

func (packet* UtilPacket) BuildPacket(bytes []byte, reuseBuf bool) (_interface.BinaryPacket, error){
	packetSize, ok := packet.ParsePacketHeader(bytes)
	if !ok {
		return nil, BuildPacketError
	}

	if uint32(len(bytes)) < packetSize {
		return nil, BuildPacketError
	}

	if reuseBuf {
		packet = new(UtilPacket)
		packet.content = bytes
		packet.readIndex = packetHeaderSize
		packet.writeIndex = uint32(len(bytes))
		packet.contentCap = packet.writeIndex
		packet.seqNo = (*uint32)(unsafe.Pointer(&packet.content[seqOffset]))
		packet.op = (*uint16)(unsafe.Pointer(unsafe.Pointer(&packet.content[opOffset])))
		packet.size = (*uint32)(unsafe.Pointer(unsafe.Pointer(&packet.content[sizeOffset])))
		packet.SetSize(packet.writeIndex)
		return packet, nil
	}else{
		packet = new(UtilPacket)
		packet.content = make([]byte, packetSize, packetSize)
		copy(packet.content, bytes[0:packetSize])
		packet.readIndex = packetHeaderSize
		packet.writeIndex = packetSize
		packet.contentCap = packet.writeIndex
		packet.seqNo = (*uint32)(unsafe.Pointer(&packet.content[seqOffset]))
		packet.op = (*uint16)(unsafe.Pointer(unsafe.Pointer(&packet.content[opOffset])))
		packet.size = (*uint32)(unsafe.Pointer(unsafe.Pointer(&packet.content[sizeOffset])))
		packet.SetSize(packet.writeIndex)
		return packet, nil
	}
}


func (packet *UtilPacket) ParsePacketHeader(header []byte) (uint32, bool) {
	if header == nil || uint32(len(header)) < packetHeaderSize {
		return 0, false
	}


	size := *(*uint32)(unsafe.Pointer(&header[sizeOffset]))
	if util.IsLittleEnd() {
		size = binaryop.SwapUint32(size)
	}
	return size, true
}

func (packet *UtilPacket) GetReadIndex() uint32 {
	return packet.readIndex
}

func (packet *UtilPacket) GetWriteIndex() uint32 {
	return packet.writeIndex
}

func (packet *UtilPacket) GetContent() []byte {
	x := (uintptr)(unsafe.Pointer(&packet.content[0]))
	h := [3]uintptr{x, uintptr(packet.writeIndex), uintptr(packet.writeIndex)}
	buf := *(*[]byte)(unsafe.Pointer(&h))
	return buf
}

func (packet *UtilPacket) Reset(op uint16) {
	packet.readIndex = packetHeaderSize
	packet.writeIndex = packetHeaderSize
	packet.SetSize(packet.writeIndex)
	if util.IsLittleEnd() {
		*packet.op = binaryop.SwapUint16(op)
	} else {
		*packet.op = op
	}
}

func (packet *UtilPacket) SetSeqNo(sNo uint32) {
	if util.IsLittleEnd() {
		*packet.seqNo = binaryop.SwapUint32(sNo)
	} else {
		*packet.seqNo = sNo
	}
}

func (packet *UtilPacket) SetSize(size uint32) {
	if util.IsLittleEnd() {
		*packet.size = binaryop.SwapUint32(size)
	} else {
		*packet.size = size
	}
}

func (packet *UtilPacket) SetOp(op uint16) {
	if util.IsLittleEnd() {
		*packet.op = binaryop.SwapUint16(op)
	} else {
		*packet.op = op
	}
}

func (packet *UtilPacket) checkPacketSize(len uint32) bool {
	if packet.writeIndex+len > maxContentSize {
		return false
	}

	if packet.writeIndex+len >= packet.contentCap {
		nLen := 2 * (packet.contentCap + len)
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

func (packet *UtilPacket) GetOpcode() uint16 {
	if util.IsLittleEnd() {
		return binaryop.SwapUint16(*packet.op)
	} else {
		return *packet.op
	}
}

func (packet *UtilPacket) GetSeqNo() uint32 {
	if util.IsLittleEnd() {
		return binaryop.SwapUint32(*packet.seqNo)
	} else {
		return *packet.seqNo
	}
}

func (packet *UtilPacket) GetSize() uint32 {
	if util.IsLittleEnd() {
		return binaryop.SwapUint32(*packet.size)
	} else {
		return *packet.size
	}
}

var err_input_buf_size = errors.New("input byte lenght < size")

func (packet *UtilPacket) WriteBytes(b []byte, size int) (uint32, error) {
	if size >= 0x01000000 {
		return packet.writeIndex, MaxSizeError
	}

	if len(b) < int(size) {
		return 0, err_input_buf_size
	}

	if !packet.checkPacketSize(uint32(size)) {
		return packet.writeIndex, MaxSizeError
	}

	copy(packet.content[packet.writeIndex:], b[0:size])
	packet.writeIndex += uint32(size)
	packet.SetSize(packet.writeIndex)
	return packet.writeIndex, nil
}

func (packet *UtilPacket) WriteString(str string) (uint32, error) {
	size := len(str)
	if size >= 0x01000000 {
		return packet.writeIndex, MaxSizeError
	}

	if !packet.checkPacketSize(uint32(size) + 3) {
		return packet.writeIndex, MaxSizeError
	}

	bb := [4]byte{}
	binaryop.Uint32ToBytes(uint32(size), &bb)
	for i := 1; i < 4; i++ {
		packet.content[packet.writeIndex+uint32(i-1)] = bb[i]
	}
	packet.writeIndex += 3

	b := binaryop.String2Byte(str)
	copy(packet.content[packet.writeIndex:], b[0:])
	packet.writeIndex += uint32(len(str))
	packet.SetSize(packet.writeIndex)
	return packet.writeIndex, nil
}

func (packet *UtilPacket) ReadString() string {
	if packet.readIndex+3 <= packet.contentCap+packetHeaderSize {
		b := [4]byte{0x00,}
		copy(b[1:], packet.content[packet.readIndex:])
		packet.readIndex += 3
		l := binaryop.BytesToUint32(b)
		if packet.readIndex+l < packet.contentCap+packetHeaderSize {
			str := string(packet.content[packet.readIndex : packet.readIndex+l])
			packet.readIndex += l
			return str
		}

	}
	return ""
}

func (packet *UtilPacket) WriteByte(b byte, wIndex ...uint32) (uint32, error) {
	writeIndex := packet.writeIndex
	if wIndex != nil {
		writeIndex = wIndex[0]
		if writeIndex >= packet.contentCap {
			return packet.writeIndex, MaxSizeError
		}
		packet.content[writeIndex] = b
	} else {
		if !packet.checkPacketSize(1) {
			return packet.writeIndex, MaxSizeError
		}
		packet.content[writeIndex] = b
		packet.writeIndex += 1
		packet.SetSize(packet.writeIndex)
	}
	return packet.writeIndex, nil
}

func getIntTypeSize(num interface{}) uint32 {
	switch num.(type) {
	case int32, uint32:
		return 4
	case int16, uint16:
		return 2
	case int8, uint8:
		return 1
	case int, uint, int64, uint64:
		return 8
	default:
		return 0
	}
}

func (packet *UtilPacket) WriteNum(num interface{}, wIndex ...uint32) (uint32, error) {
	writeIndex := packet.writeIndex
	if wIndex != nil {
		writeIndex = wIndex[0]
		if writeIndex >= packet.contentCap {
			return packet.writeIndex, MaxSizeError
		}
	}

	l := getIntTypeSize(num)
	if l == 0 {
		return packet.writeIndex, NumTypeError
	}

	if wIndex != nil {
		if writeIndex+l >= packet.contentCap {
			return packet.writeIndex, WriteAcessError
		}
	} else if !packet.checkPacketSize(l) {
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
		binaryop.Int16ToByte(val, &b)
		for i := 0; i < 2; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}

		if wIndex == nil {
			packet.writeIndex += 2
			packet.SetSize(packet.writeIndex)
		}
		return packet.writeIndex, nil

	case uint16:
		val, _ := num.(uint16)
		b := [2]byte{}
		binaryop.Uint16ToByte(val, &b)
		for i := 0; i < 2; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 2
			packet.SetSize(packet.writeIndex)
		}
		return packet.writeIndex, nil
	case int32:
		val, _ := num.(int32)
		b := [4]byte{}
		binaryop.Int32ToBytes(val, &b)
		for i := 0; i < 4; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 4
			packet.SetSize(packet.writeIndex)
		}
		return packet.writeIndex, nil
	case uint32:
		val, _ := num.(uint32)
		b := [4]byte{}
		binaryop.Uint32ToBytes(val, &b)
		for i := 0; i < 4; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 4
			packet.SetSize(packet.writeIndex)
		}
		return packet.writeIndex, nil
	case int64:
		b := [8]byte{}
		val, _ := num.(int64)
		binaryop.Int64ToBytes(val, &b)
		for i := 0; i < 8; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 8
			packet.SetSize(packet.writeIndex)
		}
		return packet.writeIndex, nil
	case uint64:
		val, _ := num.(uint64)
		b := [8]byte{}
		binaryop.Uint64ToBytes(val, &b)
		for i := 0; i < 8; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 8
			packet.SetSize(packet.writeIndex)
		}
		return packet.writeIndex, nil
	case int:
		b := [8]byte{}
		val, _ := num.(int)
		binaryop.Int64ToBytes(int64(val), &b)
		for i := 0; i < 8; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 8
			packet.SetSize(packet.writeIndex)
		}
		return packet.writeIndex, nil
	case uint:
		b := [8]byte{}
		val, _ := num.(uint)
		binaryop.Int64ToBytes(int64(val), &b)
		for i := 0; i < 8; i++ {
			packet.content[writeIndex+uint32(i)] = b[i]
		}
		if wIndex == nil {
			packet.writeIndex += 8
			packet.SetSize(packet.writeIndex)
		}
		return packet.writeIndex, nil
	default:
		return packet.writeIndex, NumTypeError
	}
}

func (packet *UtilPacket) ReadInt8() int8 {
	if packet.readIndex+1 <= packet.contentCap {
		packet.readIndex += 1
		return int8(packet.content[packet.readIndex])
	}
	packet.readIndex = packet.contentCap
	return 0
}

func (packet *UtilPacket) ReadUint8() uint8 {
	return uint8(packet.ReadInt8())
}

func (packet *UtilPacket) ReadInt16() int16 {
	if packet.readIndex+2 <= packet.contentCap {
		ret := *(*int16)(unsafe.Pointer(&packet.content[packet.readIndex]))
		packet.readIndex += 2
		if util.IsLittleEnd() {
			return binaryop.SwapInt16(ret)
		}
		return ret
	}
	packet.readIndex = packet.contentCap
	return 0
}

func (packet *UtilPacket) ReadUint16() uint16 {
	return uint16(packet.ReadInt16())
}

func (packet *UtilPacket) ReadInt32() int32 {
	if packet.readIndex+4 <= packet.contentCap {
		ret := *(*int32)(unsafe.Pointer(&packet.content[packet.readIndex]))
		packet.readIndex += 4
		if util.IsLittleEnd() {
			return binaryop.SwapInt32(ret)
		}
		return ret
	}
	packet.readIndex = packet.contentCap
	return 0
}

func (packet *UtilPacket) ReadUint32() uint32 {
	return uint32(packet.ReadInt32())
}

func (packet *UtilPacket) ReadInt64() int64 {
	if packet.readIndex+8 <= packet.contentCap {
		ret := *(*int64)(unsafe.Pointer(&packet.content[packet.readIndex]))
		packet.readIndex += 8

		if util.IsLittleEnd() {
			return binaryop.SwapInt64(ret)
		}
		return ret
	}
	packet.readIndex = packet.contentCap
	return 0
}

func (packet *UtilPacket) ReadUInt64() uint64 {
	return uint64(packet.ReadInt64())
}

func (packet *UtilPacket) ReadInt() int {
	return int(packet.ReadInt64())
}

func (packet *UtilPacket) ReadUint() uint {
	return uint(packet.ReadInt64())
}

func (packet *UtilPacket) WriteBool(b bool) (uint32, error) {
	bb := byte(0)
	if b {
		bb = 1
	} else {
		bb = 0
	}
	return packet.WriteByte(bb)
}

func (packet *UtilPacket) ReadBool() bool {
	val := packet.ReadUint8()
	return val != 0
}
