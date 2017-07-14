package api

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/gzip"
	"net/http"
	"fmt"
	"encoding/json"
	"strconv"
	"os"
	"path/filepath"
	"strings"
	"path"
	"reflect"
)

type WebServer struct {
	worker *webapi
	Handler http.Handler
}

type Result struct {
	Success bool
	Data interface{}
	Error string
}

const (
	configfilename = "all.config"
)

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}

	return strings.Replace(dir, "\\", "/", -1)
}

func getConfigFilePath() string {
	return path.Join(getCurrentDirectory(), configfilename)
}

func NewWebServer() *WebServer {
	server := new(WebServer)
	server.worker = new(webapi)
	if config, err := load(getConfigFilePath()); err == nil {
		println(config.Show())
		server.worker.Init(config)
	} else {
		println("failed to load config")
	}

	m := martini.New()
	m.Use(martini.Recovery())
	m.Use(render.Renderer())
	m.Use(func(c martini.Context, w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != "/" {
			gzip.All()
		}
	})
	m.Use(func(c martini.Context, w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != "/" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
	})

	r := martini.NewRouter()
	r.Get("/", func(r render.Render) {
		r.Text(200, "welcome to redis-group-api")
	})

	r.Group("/group/:groupx", func(r martini.Router){
		r.Get("/get/:key", server.Get)
		r.Get("/set/:key/:value", server.Set)
		r.Get("/setex/:key/:value/:expire", server.SetEx)
		r.Get("/echo/:key", server.Echo)
	})


	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)

	server.Handler = m
	println("server setup")
	return server
}

func (s *WebServer) Exit() {
	if s.worker != nil {
		s.worker.Close()
	}
}

func (s *WebServer) Get(param martini.Params) string {
	return s.process(s.worker.Get, param, "key")
}

func (s *WebServer) Set(param martini.Params) string {
	return s.process(s.worker.Set, param, "key", "value")
}

func (s *WebServer) SetEx(param martini.Params) string {
	return s.process(s.worker.SetExp, param, "key", "value", "expire")
}

func (s *WebServer) Echo(param martini.Params) string {
	groupname := param["groupx"]
	key := param["key"]
	result := Result{}
	result.Success = true
	result.Data = fmt.Sprintf("group:[%s], key:[%s]", groupname, key)

	return vtostring(result)
}

func (s *WebServer) process(f interface{}, params martini.Params, paramnames ...string) string {
	gpkey := getGroupKey(params)
	rparams := make([]string, len(paramnames))

	for i, name := range paramnames {
		rparams[i] = params[name]
	}

	result := Result{}
	if val, err := call(f, gpkey, rparams); err != nil {
		result.Success = false
		result.Error = err.Error()
	} else {
		result.Success = true
		result.Data = val
	}
	return vtostring(result)
}

func getGroupKey(params martini.Params) GroupKey {
	groupname := params["groupx"]
	groupid := -1
	if groupid2, err := strconv.Atoi(groupname); err == nil {
		groupid = groupid2
	}

	return GroupKey{GroupName:groupname, GroupId:groupid}
}

func call(f interface{}, gpkey GroupKey, params []string) (interface{}, error) {
	t := reflect.TypeOf(f)
	in := make([]reflect.Value, t.NumIn())
	in[0] = reflect.ValueOf(gpkey)
	for i, p := range params {
		in[i+1] = reflect.ValueOf(p)
	}
	vs := reflect.ValueOf(f).Call(in)


	errv := vs[1].Interface()
	var err error
	if errv != nil {
		err = errv.(error)
	} else {
		err = nil
	}

	return vs[0].Interface(), err
}

func vtostring(v interface{}) string {
	if res, err := serialize(v); err != nil {
		return "{\"success\":false, \"Error\":\"serialize result error\"}"
	} else {
		return string(res)
	}
}

func serialize(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "    ")
}