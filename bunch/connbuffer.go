package bunch

import "net"

type connBuffer struct {
	BuffReader *redisReader
	BuffWriter *redisWriter
}

func NewConnBuffer(conn net.Conn, bufferSize int) *connBuffer{
	buffer := &connBuffer{}
	buffer.BuffReader = &redisReader{reader:NewBufferReader(conn, bufferSize)}
	buffer.BuffWriter = &redisWriter{writer:NewBufferWriter(conn, bufferSize)}
	return buffer
}
