package bunch

import (
	"github.com/simazhao/go-redis-group/config"
	"github.com/simazhao/go-redis-group/datamodel"
	"github.com/simazhao/go-redis-group/log"
	"net"
	"errors"
	"time"
)

type redisConn struct {
	address string
	conn net.Conn
	config config.ClientConfig
	state int
	requests chan *datamodel.Request
}

const (
	cold = 0
	running = 1
	stopped = 2
)

func NewRedisConn(address string, config config.ClientConfig) *redisConn{
	conn := &redisConn{config:config,address:address}
	conn.requests = make(chan *datamodel.Request, config.RequsetChanLength)

	conn.Run()

	return conn
}

func (dr *redisConn) Run() (err error) {
	if dr.state == running {
		return nil
	}

	dr.state = running

	go dr.tryWork()

	return nil
}

func (dr *redisConn) IsRunning() bool {
	return dr.state == running
}

func (dr *redisConn) tryWork()  {
	for true {
		shouldDelay := false
		if c, err := dr.connect(); err != nil {
			shouldDelay = true
		} else {
			dr.conn = c
			if err := dr.work(c); err != nil{
				shouldDelay = true
			}
		}

		if !dr.IsRunning() {
			dr.disconnect()
			return
		}

		if shouldDelay {
			delay := time.After(time.Second * 30)
			select {
			case <-delay:
				continue
			case r, ok := <- dr.requests:
				if ok {
					r.Clips = nil
					r.Err = errors.New("redis do not connected")
					if r.Wait != nil  {
						r.Wait.Done()
					}
					log.Factory.GetLogger().WarnFormat( "request %s can not be handled", r)
				}
			}
		}
	}
}

func (dr *redisConn) work(conn net.Conn) error {
	defer func() {
		if len(dr.requests) > 0{
			for request := range dr.requests {
				log.Factory.GetLogger().WarnFormat( "request %s is abandoned", request)
			}
		}
	}()

	connbuffer := NewConnBuffer(conn, dr.config.BufferSize)
	readChan := make(chan *datamodel.Request, 1024)
	go dr.tryRead(readChan, connbuffer.BuffReader)


	for request := range dr.requests {
		if !dr.IsRunning() {
			return nil
		}

		// use same buffer to handle write
		if err := connbuffer.BuffWriter.WriteRequest(request); err != nil{
			log.Factory.GetLogger().ErrorFormat("handleRequest error: %s", err.Error())
			dr.setResponse(request, nil, err)
			return err
		} else {
		}

		log.Factory.GetLogger().Info("put request to read chan")
		readChan <- request
	}

	return errors.New("work abort")
}

func (dr *redisConn) tryRead(readChan chan *datamodel.Request, reader *redisReader) error {
	defer func() {
		if len(readChan) > 0 {
			for request := range readChan {
				log.Factory.GetLogger().ErrorFormat( "request %s is abandoned", request)
			}
		}
	}()

	for request := range readChan {
		if !dr.IsRunning() {
			return dr.setResponse(request, nil, errors.New("redisConn is not running"))
		}

		log.Factory.GetLogger().Info("read request from readChan")
		// use same buffer to handle read
		if response, err := reader.ReadResponse(); err != nil {
			dr.setResponse(request, nil, err)
		} else {
			dr.setResponse(request, response, nil)
		}
	}

	return errors.New("read abort")
}

func (dr *redisConn) setResponse(request *datamodel.Request, res []*datamodel.RequestClip, err error) error {
	request.Clips, request.Err = res, err

	if request.Wait != nil {
		request.Wait.Done()
	}

	return err
}

func (dr *redisConn) Stop() error {
	if dr.state == stopped {
		return nil
	}

	defer func() {
		dr.requests <- quit
	}()

	dr.state = stopped
	return nil
}

func (dr *redisConn) connect() (net.Conn, error){
	if c, err := net.DialTimeout("tcp", dr.address, dr.config.ConnectionTimeout); err != nil {
		return nil, err
	} else {
		return c, nil
	}
}

var quit = &datamodel.Request{}

func (dr *redisConn) disconnect() error {
	return dr.conn.Close()
}

func (dr *redisConn) Put(request *datamodel.Request) error {
	if !dr.IsRunning() {
		return errors.New("redisConn is not running")
	}

	if request.Wait != nil {
		request.Wait.Add(1)
	}

	dr.requests <- request

	return nil
}