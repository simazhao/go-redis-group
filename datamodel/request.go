package datamodel

import "sync"

type SegmentType byte

const (
	SegString    SegmentType = '+'
	SegError     SegmentType = '-'
	SegInt       SegmentType = ':'
	SegBulkBytes SegmentType = '$'
	SegArray     SegmentType = '*'
)


type Request struct{
	Clips []*RequestClip
	Wait *sync.WaitGroup
	Database int

	TimeStamp int64
	Err error
}

type RequestClip struct {
	Type SegmentType
	Value []byte
}