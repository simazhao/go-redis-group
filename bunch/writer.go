package bunch

import (
	"net"
	"strconv"
	"errors"
	"io"
	"fmt"
	"github.com/simazhao/go-redis-group/datamodel"
)


type redisWriter struct {
	writer *BufferWriter
}

func (e *redisWriter) WriteRequest(r *datamodel.Request) (err error) {
	if err = e.writeTotalLength(len(r.Clips)); err != nil {
		return
	}

	for _, clip := range r.Clips {
		if err = e.writeRequestClip(clip); err != nil {
			return
		}
	}

	_, err = e.writer.Flush()

	return
}

func (e *redisWriter) writeTotalLength(length int) (err error){
	if _, err = e.writer.WriteByte(byte(datamodel.SegArray)); err != nil {
		return
	}

	if err = e.writeInt(length); err != nil {
		return
	}

	if err = e.writeEndLine(); err != nil{
		return
	}

	return
}

func (e* redisWriter) writeInt(n int) error {
	return e.WriteString(strconv.Itoa(n))
}

func (e *redisWriter) WriteString(str string) error {
	_, err := e.writer.Write([]byte(str))

	return err
}

const (
	CR = '\r'
	LF = '\n'
	CRLF = "\r\n"
)

var crlfBytes = []byte(CRLF)

func (e *redisWriter) writeEndLine() (err error) {
	_, err = e.writer.Write(crlfBytes)

	return
}

func (e *redisWriter) writeRequestClip(r *datamodel.RequestClip) (err error) {
	switch r.Type {
	case datamodel.SegBulkBytes:
		return e.writeBulkRequest(r)
	case datamodel.SegInt:
	case datamodel.SegString:
		return e.WriteText(r)
	default:
		return errors.New(fmt.Sprint("do not supported request type :%s" , r.Type))
	}

	return
}

func (e *redisWriter) writeBulkRequest(r *datamodel.RequestClip) (err error){
	if err = e.writeRequstClipLength(len(r.Value)); err != nil {
		return
	}

	if _, err = e.writer.Write(r.Value); err != nil {
		return
	}

	if err = e.writeEndLine(); err != nil{
		return
	}

	return
}

func (e *redisWriter) writeRequstClipLength(length int) (err error) {
	if _, err = e.writer.WriteByte(byte(datamodel.SegBulkBytes)); err != nil {
		return
	}

	if err = e.writeInt(length); err != nil {
		return
	}

	if err = e.writeEndLine(); err != nil{
		return
	}

	return
}

func (e *redisWriter) WriteText(r* datamodel.RequestClip) (err error) {
	if _, err = e.writer.Write(r.Value); err != nil {
		return
	}

	return e.writeEndLine()
}


type BufferWriter struct {
	Writer *connWriter
	buf []byte
	wpos int
}

func NewBufferWriter(conn net.Conn, bufferSize int) *BufferWriter {
	buffer := BufferWriter{Writer:&connWriter{Conn:conn}}
	buffer.buf = make([]byte, bufferSize)
	return &buffer
}

func (bw *BufferWriter) capacity() int {
	return len(bw.buf)
}

func (bw *BufferWriter) leftroom() int {
	return bw.capacity() - bw.wpos
}

func (bw *BufferWriter) WriteByte(s byte) (n int, err error) {
	if bw.leftroom() == 0 {
		if _, err = bw.Flush(); err != nil {
			return
		}
	}

	bw.buf[bw.wpos] = s
	bw.wpos += 1

	return 1, nil
}

func (bw *BufferWriter) Write(bytes []byte) (n int, err error) {
	for err == nil && len(bytes) > bw.leftroom(){
		nc := copy(bw.buf[bw.wpos:], bytes)
		bw.wpos += nc
		_, err = bw.Flush()

		n, bytes = n+nc, bytes[nc:]
	}

	if err != nil {
		return
	}

	if len(bytes) == 0 {
		err = errors.New("incorrect bytes length")
		return
	}

	nc := copy(bw.buf[bw.wpos:], bytes)
	n += nc
	bw.wpos += nc
	return
}

func (bw *BufferWriter) Flush() (n int, err error) {
	if bw.wpos == 0 {
		return
	}

	n, err = bw.Writer.Write(bw.buf[:bw.wpos])

	if err == nil {
		if n < bw.wpos {
			err = io.ErrShortWrite
		} else {
			bw.wpos = 0
		}
	}

	return
}


type connWriter struct {
	Conn net.Conn
}

func (w *connWriter) Write(bytes []byte) (int, error) {
	if w.Conn == nil {
		return 0, errors.New("no connection")
	}

	return w.Conn.Write(bytes)
}