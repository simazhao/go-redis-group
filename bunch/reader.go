package bunch

import (
	"net"
	//"go-redis-driver/pkg/log"
	"strconv"
	"errors"
	"bytes"
	"fmt"
	"time"
	"github.com/simazhao/go-redis-group/datamodel"
)

type redisReader struct {
	reader *BufferReader
}

func (r *redisReader) ReadResponse() ([]*datamodel.RequestClip, error) {
	response := make([]*datamodel.RequestClip, 0)

	if err := r.readResponse(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func (r *redisReader) readResponse(clips *[]*datamodel.RequestClip) error {
	if head, err := r.readPrefix(clips); err != nil {
		return err
	} else {
		return r.readMore(datamodel.SegmentType(head), clips)
	}

	return nil
}

func (r *redisReader) readPrefix(clips *[]*datamodel.RequestClip) (byte, error) {
	if head, err := r.reader.ReadByte(); err != nil {
		return 0, err
	} else {
		if head != byte(datamodel.SegArray) {
			clip := &datamodel.RequestClip{Type:datamodel.SegmentType(head)}
			*clips = append(*clips, clip)
		}

		return head, nil
	}
}

func (r *redisReader) readMore(segtype datamodel.SegmentType, clips *[]*datamodel.RequestClip) error {
	switch segtype {
	case datamodel.SegString:
		return r.readString(clips)
	case datamodel.SegError:
		return r.readError(clips)
	case datamodel.SegInt:
		return r.readString(clips)
	case datamodel.SegBulkBytes:
		return r.readBulk(clips)
	case datamodel.SegArray:
		return r.readArray(clips)
	default:
		return errors.New(fmt.Sprintf("not support %c", segtype))
	}
}

func (r *redisReader) readBulk(clips *[]*datamodel.RequestClip) error {
	if lengthstr, err := r.reader.ReadLine(); err != nil {
		return err
	} else if length, err := strconv.Atoi(string(lengthstr[:len(lengthstr)-2])); err != nil {
		return errors.New(fmt.Sprintf("bulk data length is illegal %d %d %d", len(lengthstr), r.reader.rpos, r.reader.wpos))
	} else if length < 0 {
		clip := (*clips)[len(*clips) - 1]
		clip.Value = make([]byte, 0)
		return nil
	} else if bytes, err := r.reader.ReadLine(); err != nil {
		return err
	} else if len(bytes) != length + 2 {
		return errors.New("data length is incorrect")
	} else {
		clip := (*clips)[len(*clips) - 1]
		clip.Value = make([]byte, len(bytes) - 2)
		copy(clip.Value, bytes)
		return nil
	}
}

func (r *redisReader) readArray(clips *[]*datamodel.RequestClip) error {
	if lengthstr, err := r.reader.ReadLine(); err != nil {
		return err
	} else if length, err := strconv.Atoi(string(lengthstr[:len(lengthstr)-2])); err != nil || length <= 0{
		return errors.New("data length is illegal")
	} else {
		for i:=0;i<length;i++ {
			if err := r.readResponse(clips); err != nil{
				return err
			}
		}
	}

	return nil
}

func (r* redisReader) readError(clips *[]*datamodel.RequestClip) error {
	if _, err := r.reader.ReadLine(); err != nil {
		return err
	} else {
		return errors.New("redis return error")
	}
}

func (r* redisReader) readString(clips *[]*datamodel.RequestClip) error {
	if bytes, err := r.reader.ReadLine(); err != nil {
		return err
	} else {
		clip := (*clips)[len(*clips) - 1]
		clip.Value = make([]byte, len(bytes) - 2)
		copy(clip.Value, bytes)
		return nil
	}
}

type BufferReader struct{
	Reader *connReader
	buf []byte
	rpos int
	wpos int
}

func (br *BufferReader) leftroom() int {
	return len(br.buf) - br.wpos
}

func (br *BufferReader) useful() int{
	return br.wpos - br.rpos
}

func NewBufferReader(conn net.Conn, bufferSize int) *BufferReader {
	buffer := &BufferReader{Reader:&connReader{Conn:conn}}
	buffer.buf = make([]byte, bufferSize)
	return buffer
}

func (br *BufferReader) ReadLine() ([]byte, error) {
	line,hasnext,err := br.readLine()

	if err != nil {
		return nil, err
	}

	if !hasnext {
		return line, nil
	}

	ret := make([]byte, len(line))
	copy(ret, line)

	for {
		line,hasnext,err = br.readLine()
		if err != nil {
			return nil, err
		}

		ret = concatarray(ret, line)

		if !hasnext {
			return ret, nil
		}
	}
}

func concatarray(left []byte, right []byte) []byte {
	newarray := make([]byte, len(left) + len(right))

	if newarray == nil {
		panic("can not alloc more memory")
	}

	i := 0
	for _,b := range left {
		newarray[i] = b
		i++
	}

	for _,b := range right {
		newarray[i] = b
		i++
	}

	return newarray
}

func (br *BufferReader) readLine() (line []byte, hasnext bool, err error) {
	for {
		rpos := br.rpos
		wpos := br.wpos

		if n := containLine(br.buf[br.rpos:br.wpos]); n >= 0 {
			br.rpos = rpos + n + 1

			return br.buf[rpos:br.rpos], false, nil
		}

		if br.useful() >= len(br.buf) {
			lastbyte := br.buf[len(br.buf) - 1]
			if lastbyte == CR {
				br.rpos = br.wpos - 1
				wpos = br.rpos
			} else {
				br.rpos, br.wpos = 0, 0
			}

			return br.buf[rpos:wpos], true, nil
		}

		if rpos > 0 {
			copy(br.buf, br.buf[br.rpos:br.wpos])
			br.rpos, br.wpos = 0, br.wpos - br.rpos
		}

		if n, err := br.Reader.Read(br.buf[br.wpos:]); err != nil {
			return nil, false, err
		} else {
			br.wpos += n
		}
	}
}

func containLine(data []byte) int {
	if len(data) < 2 {
		return -1
	}

	index := bytes.IndexByte(data, LF)
	if index > 0 && data[index - 1]  == CR {
		return index
	}

	return -1
}

func (br *BufferReader) ReadByte() (byte, error) {
	rpos := br.rpos
	if br.useful() > 0 {
		br.rpos += 1
		return br.buf[rpos], nil
	}

	if br.leftroom() == 0 {
		rpos, br.rpos, br.wpos = 0, 0, 0
	}

	n, err := br.Reader.Read(br.buf[br.wpos:])

	if err != nil {
		return 0, err
	}

	br.wpos += n
	br.rpos += 1

	return br.buf[rpos], nil
}

type connReader struct{
	Conn net.Conn
}

func (c *connReader) Read(bytes []byte) (int, error) {
	c.Conn.SetReadDeadline(time.Now().Add(time.Second*5))
	return c.Conn.Read(bytes)
}
