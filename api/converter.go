package api

import (
	"encoding/json"
	"sync"
	"time"
	"fmt"
	"github.com/simazhao/go-redis-group/datamodel"
	"reflect"
	"strconv"
	"github.com/CodisLabs/codis/pkg/utils/errors"
)

func covertType(t reflect.Type, v string) (reflect.Value, error) {
	ttname := t.Name()

	switch ttname {
	case "string":
		return reflect.ValueOf(v), nil
	case "int":
		if n, err := strconv.Atoi(v); err != nil {
			return reflect.Value{}, errors.New("incorrect param type")
		} else {
			return reflect.ValueOf(n), nil
		}
	case "Duration":
		if n, err := strconv.ParseInt(v, 10, 0); err != nil {
			return reflect.Value{}, errors.New("incorrect param type")
		} else {
			return reflect.ValueOf(time.Duration(n)), nil
		}
	}

	return reflect.Value{}, errors.New("unknown param")
}

func newRequest() (*datamodel.Request) {
	return &datamodel.Request{Wait:&sync.WaitGroup{}, TimeStamp:time.Now().UnixNano()}
}

func getSetBytes(val interface{}) ([]byte, error) {
	var valbytes []byte
	var err error
	if v, ok := val.(string); ok {
		valbytes = []byte(v)
	} else if valbytes, err = json.Marshal(val); err != nil {
		return nil, err
	}

	return valbytes, nil
}


func convertSetKeyValue(key string, val interface{}) (*datamodel.Request, error){
	if valbytes, err := getSetBytes(val); err != nil {
		return nil, err
	} else {
		request := newRequest()
		request.Clips = makeRequestClips(convertToBytes(datamodel.SET), convertToBytes(key), valbytes)
		return request, nil
	}
}

func newDuration(hours int64, minutes int64, seconds int64) time.Duration {
	return time.Duration(hours * int64(time.Hour) + minutes * int64(time.Minute) + seconds * int64(time.Second))
}

func convertSetKeyValueExpire(key string, val interface{}, dur time.Duration) (*datamodel.Request, error){
	if valbytes, err := getSetBytes(val); err != nil {
		return nil, err
	} else {
		request := newRequest()
		seconds := int64(dur/time.Second)
		request.Clips = makeRequestClips(convertToBytes(datamodel.SETEX), convertToBytes(key), convertIntToBytes(seconds), valbytes)
		return request, nil
	}
}

func convertGetKeyValue(key string) (*datamodel.Request, error) {
	request := newRequest()
	request.Clips = makeRequestClips(convertToBytes(datamodel.GET), convertToBytes(key))
	return request, nil
}

func convertMGetKeyValue(keys []string) (*datamodel.Request, error) {
	request := newRequest()
	request.Clips = make([]*datamodel.RequestClip, len(keys) + 1)
	request.Clips[0] = newRequestClip(convertToBytes(datamodel.MGET))

	for i, key := range keys {
		request.Clips[i+1] = newRequestClip(convertToBytes(key))
	}

	return request, nil
}

func convertMSetKeyValue(keyvals map[string]interface{}) (*datamodel.Request, error) {
	request := newRequest()
	request.Clips = make([]*datamodel.RequestClip, len(keyvals)*2 + 1)
	request.Clips[0] = newRequestClip(convertToBytes(datamodel.MSET))

	i := 1
	for key, val := range keyvals {
		request.Clips[i] = newRequestClip(convertToBytes(key))

		if valbytes, err := json.Marshal(val); err != nil {
			return nil, err
		} else {
			request.Clips[i+1] = newRequestClip(valbytes)
		}

		i += 2
	}

	return request, nil
}

func convertExpireAt(key string, dt time.Time) (*datamodel.Request, error) {
	request := newRequest()

	request.Clips = makeRequestClips(convertToBytes(datamodel.EXPIREAT), convertToBytes(key), convertIntToBytes(dt.Unix()))

	return request, nil
}

func convertIntToBytes(p int64) []byte {
	return []byte(fmt.Sprintf("%d", p))
}

func convertToBytes(p string) []byte {
	return []byte(p)
}

func makeRequestClips(params ...[]byte) ([]*datamodel.RequestClip){
	clips := make([]*datamodel.RequestClip, len(params))
	for i,param := range params {
		clips[i] = newRequestClip(param)
	}

	return clips
}

func newRequestClip(param []byte) *datamodel.RequestClip {
	return &datamodel.RequestClip{Type:datamodel.SegBulkBytes, Value:param}
}