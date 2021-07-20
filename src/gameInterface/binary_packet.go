package gameInterface

type BinaryPacket interface {
	//safe ...interface{} first param must bool, false, unsafe, true safe, default safe

	HeaderSize() uint32
	GetHeader(safe ...interface{}) []byte

	BodySize() uint32
	GetBody(safe ...interface{}) []byte

	PacketSize() uint32
	GetPacket(safe ...interface{}) []byte  //header + body

	ParsePacketHeader(bytes []byte) (uint32, bool) //返回packet size, headersize + bodysize
	BuildPacket(bytes []byte, reuseBuf bool) (BinaryPacket, error) //通过bytes直接构建packet, reuserBuf会复用bytes内存
}
